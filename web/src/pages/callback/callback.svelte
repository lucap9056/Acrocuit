<script lang="ts">
  import { onMount } from "@/lib/lifecycle";
  import { notifyLoginCallback } from "@/lib/login-popup";

  onMount(() => {
    notifyLoginCallback(location.search);
    window.close();
  });
</script>

<main>
  <h2>Sign-in received</h2>
  <p class="muted">You can close this window and return to Acrocuit.</p>
</main>

<style>
  main {
    display: flex;
    flex: 1;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    padding: 16px;
    text-align: center;
  }
</style>
