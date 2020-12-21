<template>
  <div>
    <body>
      <Navbar />
      <section id="wrapper" class="section" style="padding-top: 0">
        <div class="tile is-ancestor">
          <div class="tile is-vertical is-8">
            <div class="tile is-parent">
              <article class="tile is-child">
                <p class="subtitle" style="padding-top: 0">
                  My name is Ze Pedro Resende, I'm a Software Engineer at
                  <a href="https://finiam.com">Finiam</a>. <br />
                </p>
                <p class="content">
                  I'm interested in Audio, Distributed Systems, Data Exploration
                  and Visualization, High Performance Computing.<br />
                  Right now i'm mostly programming in Ruby, Elixir, Rust and
                  Typescript.
                </p>
              </article>
            </div>
          </div>
        </div>
      </section>
      <Footer />
    </body>
  </div>
</template>

<script lang="ts">
import Vue from 'vue'

export default Vue.extend({})
</script>

<style>
.body {
  height: 100%;
}

#wrapper {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}
</style>
