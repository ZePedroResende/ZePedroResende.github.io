<script context="module" lang="ts">
	export const prerender = true;
</script>

<template>
	<section class="section is-flex is-flex-direction-column">
		<div class="container">
			<div class="tile is-parent">
				<article class="tile is-child">
					<p class="subtitle has-text-centered">
						I'm a Software Engineer at
						<a href="https://finiam.com">Finiam</a>. <br />
					</p>
					<p class="content">
						I'm interested in Audio, Distributed Systems, Data Exploration, High Performance
						Computing.<br />
					</p>

					<p class="content">I have a BSC in Software Engineering from Universidade do Minho.</p>

					<p class="content">
						I am currently in the final stages of writing my dissertation to complete my MSc in Distributed Systems and Parallel Computing.
					</p>

					<p class="content">
						Right now i'm doing Rust, Elixir, Solidity and Cairo related stuff.<br />
					</p>
				</article>
			</div>
		</div>
	</section>
</template>

<style>
	section a {
		background: black;
		color: white;
	}

	section {
		flex-direction: column;
	}
</style>
