---
title: Exponential back-off with Elixir

author: Jose Resende

date: 11-04-2023
---

# Exponential backoff with Elixir

<!--toc:start-->

- [Exponential backoff with Elixir](#exponential-backoff-with-elixir)
  - [Getting started](#getting-started)
  - [GenServer for the stores](#genserver-for-the-stores)
  - [GenServer for the allowlist](#genserver-for-the-allowlist)
  - [GenServer for polling a store](#genserver-for-polling-a-store)
  - [Conclusion](#conclusion)
  <!--toc:end-->

When we rely on external services, they may not always be available, may take a
long time to respond, or may not be able to fulfil our requests. This can be
frustrating, but if it's a one-time request and we only need to query one host,
we can simply keep trying until we get a response.

However, if we have multiple providers for the same service and need to query
them repeatedly, such as in a scenario where we're tracking prices for different
franchises of a store and maintaining a leaderboard of the best prices, we can't
rely on the same strategy of continually polling a single service. In these
cases, we need a more sophisticated approach, such as exponential backoff, to
manage the requests and handle any failures more efficiently.

In some cases, the stores that we're polling may not have the specific item
we're looking for, or they may have an unstable server or poor internet
connection. This can make it challenging to avoid wasting time on unsuccessful
requests.

So how can we prevent permanent polling, active wait, blocking code, and
scenarios like this:

> you - are you ready yet ?
>
> service - NO.
>
> you - What about now?
>
> service - ..... for the last time NO.
>
> you - And now?
>
> service - still NO.
>
> you - Hey service could you give me ...?
>
> service - Iup here you go!

## Getting started

Today I'm gonna talk about exponential back-off, and how to deal with this kind
of problem in a very simple way.

Better yet I'm gonna show you some examples in Elixir that make this whole
process a lot simpler with OTP.

Exponential backoff is like taking a deep breath and counting to 10 before
trying again when you don't succeed at something on the first try. But instead
of counting to 10, you wait a little longer each time you try again, giving
yourself and others around you a chance to catch their breath and avoid chaos.

We will use GenServers so the final result ends up in a very simple and elegant
implementation.

The premises are:

- You make the request:
  - The request went through. Great, do the next request in 5 seconds.
  - The request didn't go through, we'll try again in 25 seconds (5^2).

Now we could do this to infinity but let's do a smarter implementation with an
allowlist and blocklist for the stores that return `n` unsuccessful requests so
we don't have an infinite waiting time for the next request

- You make the request :
  - The request went through. Great, do the next request in 5 seconds.
  - The request didn't go through, we'll try again in 25 seconds (5^2) and you
    get a ticket (1).
  - The request didn't go through and you already have 3 tickets, you're going
    on the blocklist.

> Cool so how do I implement this in Elixir?

Well, it's pretty easy!

We can have this structure :

- GenServer for the Stores
- GenServer for the Allowlist
- GenServer for polling each store

## GenServer for the stores

In this, we have the main GenServer that will create the allowlist, and each of
the polling store processes.

If you don't know, a GenServer is implemented in two parts: the client API and
the server callbacks. If you want to learn more check
[this blog post on concurrency mechanisms in Elixir that I did a while back](https://blog.finiam.com/blog/genserver-agent-task).

At the public API level, we only need a way to get the prices. This is the
objective of this module, get prices.

```elixir
defmodule Project.Stores do
 def prices do
    GenServer.call(__MODULE__, :prices)
 end
 # ....
end
```

On the private callback side, we need to support the public price call (of
course) and the refresh callback to update the allow list.

The price call is gonna query the allowlist for the prices of the stores that
are currently available and answering to our queries.

```elixir
defmodule Project.Stores do
  # ...
  # private API
  def handle_call(:prices, _from, %{allowlist: allowlist} = state),
    do: {:reply, Stores.Allowlist.list(allowlist), state}

  def handle_info(:refresh, %{period: period, allowlist:   allowlist} = state) do
    case API.list_stores() do
      {:ok, stores} ->
        start_store_polling(stores, allowlist)

      error ->
        error
    end

    schedule_work(period)

    {:noreply, state}
  end
  # ...
end
```

Finally, we can have private functions like starting the polling GenServer for
each store and a function to periodically wake up the process by sending a
message to self.

This message will be caught by the handle_info on the private callback allowing
for a refresh of the stores.

```elixir
defmodule Project.Stores do
  # ...
  defp start_store_polling(stores, allowlist) do
    for store <- stores do
      Stores.Polling.start_link(store, allowlist)
    end
  end

  defp schedule_work(period), do: Process.send_after(self(), :refresh, period)
end
```

## GenServer for the allowlist

This allowlist will both keep all the current stores that are responding as well
as their prices and a quarantine list for stores that may be removed if they
keep failing to respond. With this we need the public API to support adding a
store, adding its price, removing a store, quarantining a store, and finally
listing the stores and prices.

```elixir
defmodule Project.Stores.Allowlist do
  use GenServer

  def start_link,
    do: GenServer.start_link(__MODULE__, :ok)

  def init(:ok),
    do: {:ok, %{allowlist: %{}, quarantine: MapSet.new()}}

  def add(pid, store, price),
    do: GenServer.cast(pid, {:add, store, price})

  def remove(pid, store),
    do: GenServer.cast(pid, {:remove, store})

  def quarantine(pid, store),
    do: GenServer.cast(pid, {:quarantine, store})

  def list(pid),
    do: GenServer.call(pid, :list)
  # ...
end
```

The private API is self-describing, the cast adds, removes, lists, and changes a
store from the allow list to the quarantine list. There's nothing outside of the
ordinary here.

```elixir
defmodule Project.Stores.Allowlist do
  # ...
  # private API
  def handle_cast({:add, store, price}, state) do
    quarantine = MapSet.delete(state.quarantine, store)
    allowlist = Map.put(state.allowlist, store, price)

    {:noreply, %{state | quarantine: quarantine, allowlist: allowlist}}
  end

  def handle_cast({:remove, store}, state) do
    quarantine = MapSet.delete(state.quarantine, store)
    allowlist = Map.delete(state.allowlist, store)

    {:noreply, %{state | quarantine: quarantine, allowlist: allowlist}}
  end

  def handle_cast({:quarantine, store}, state) do
    allowlist = Map.delete(state.allowlist, store)
    quarantine = MapSet.put(state.quarantine, store)

    {:noreply, %{state | quarantine: quarantine, allowlist: allowlist}}
  end

  def handle_call(:list, _from, state) do
    {:reply, Map.values(state.allowlist), state}
  end
end
```

## GenServer for polling a store

The polling store GenServer itself is where all the fun stuff happens. It keeps
track of all the requests and the number of failed attempts to communicate to
the store and loads constant values like the number of `max_attempts` allowed
and the `timeout`/`base_timeout`.

```elixir
defmodule Project.Stores.Polling do
  use GenServer

  import Project.Config, only: [config!: 2]

  alias Project.Stores.Allowlist

  def start_link(store, allowlist) do
    name = process_name(store)

    state = %{
      store: store,
      allowlist: allowlist,
      attempts: 0,
      max_attempts: config!(__MODULE__, :max_attempts),
      timeout: 2_000,
      base_timeout: 2_000
    }

    GenServer.start_link(__MODULE__, state, name: name)
  end
  # ...
end
```

On the `init` function we pass the state that we received from `start_link` and
call the private function `schedule_work` that will send a message to this
process, and make it do some work (the name is accurate).

```elixir
defmodule Project.Stores.Polling do
  # ...
  def init(state) do
    schedule_work(state.base_timeout)

    {:ok, state}
  end
  # ...
end
```

When the process receives a message to do some work it queries the store for the
current price. Depending on the result, it's handled differently. This is
especially easy with tuples in Elixir using the `:ok` and `:error` symbols on
the first element.

```elixir
defmodule Project.Stores.Polling do
  # ...
  def handle_info(:work, state) do
    case query_price(state.store) do
      {:ok, price} ->
        handle_successful_polling(state, price)

      {:error, _reason} ->
        handle_missing_polling(state)
    end
  end
  # ...
end
```

If the polling is successful we add the store and the price to the allowlist
GenServer. We will reschedule the polling with the base timeout and update the
process state to set the timeout to the base timeout with 0 retry attempts.

```elixir
defmodule Project.Stores.Polling do
  # ...
  defp handle_successful_polling(state, price) do
    Allowlist.add(state.allowlist, state.store, price)
    schedule_work(state.base_timeout)

    new_state = %{state | timeout: state.base_timeout, attempts: 0}

    {:noreply, new_state}
  end
  # ...
end
```

If the process misses the poll and it's the last attempt we are going to remove
it from the Allowlist and update the state to mark the attempts as the max
attempts.

```elixir
defmodule Project.Stores.Polling do
  # ...
  defp handle_missing_polling(
         %{attempts: attempts, max_attempts: max_attempts} = state
       )
       when attempts == max_attempts - 1 do
    Allowlist.remove(state.allowlist, state.store)

    {:stop, :normal, %{state | attempts: max_attempts}}
  end
  # ...
end
```

If it's not the last attempt and the process misses the polling we quarantine
that store, exponentiate the timeout and increase the attempt counter. After
that, we set the new state with the timeout and attempts while scheduling the
next poll with the new timeout.

```elixir
defmodule Project.Stores.Polling do
  # ...
  defp handle_missing_polling(state) do
    Allowlist.quarantine(state.allowlist, state.store)

    timeout = exponential_backoff(state.timeout)
    attempts = state.attempts + 1
    new_state = %{state | timeout: timeout, attempts: attempts}

    schedule_work(timeout)

    {:noreply, new_state}
  end
  # ...
end
```

At the end of the file, we define the helper functions for this module,
exponential backoff, schedule_work, and process name.

The exponential backoff function is a wrapper around the power function from
_Erlang_ with a round of the value, the schedule work function is a send-after
to the same process, and the process_name uses the module and store to give a
unique name to each process.

```elixir
defmodule Project.Stores.Polling do
  # ...
  defp exponential_backoff(timeout), do: :math.pow(timeout, 2) |> round()

  defp schedule_work(timeout), do: Process.send_after(self(), :work, timeout)

  defp process_name(store), do: :"#{__MODULE__}-#{store}"
end
```

## Conclusion

This was a fairly technical blog post but I truly hope you learned something
useful. Although the example presented in this post is somewhat specific, with
small changes it's possible to adapt it to your specific use case.

Hope this will help you in your next project if an exponential backoff is
needed, maybe when you are dependent on some service that rate limits requests
with [PlugAttack](https://github.com/michalmuskala/plug_attack) as described in
this [blog post](https://felt.com/blog/rate-limiting).

Thank you for reading and if you have any questions or doubts feel free to reach
out on Twitter at [@Resende_666](https://twitter.com/Resende_666).
