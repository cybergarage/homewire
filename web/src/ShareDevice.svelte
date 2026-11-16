<script lang="ts">
  import { onDestroy } from 'svelte';
  import qrcode from 'qrcode-generator';
  import { ApiError, shareNode, stopSharing } from './api';
  import type { Share } from './api';

  interface Props {
    /** The node to share. */
    node: string;
    /** The name of the device, for the explanation. */
    name: string;
    /** Whether the node is online; a node must be to open a window. */
    online: boolean;
    /** Called when the panel closes. */
    onclose: () => void;
  }

  let { node, name, online, onclose }: Props = $props();

  type Phase = 'confirm' | 'opening' | 'open' | 'closing' | 'expired';
  let phase = $state<Phase>('confirm');
  let share = $state<Share | null>(null);
  let problem = $state('');
  let now = $state(Date.now());

  const tick = setInterval(() => {
    now = Date.now();
    if (phase === 'open' && share && Date.parse(share.expiresAt) <= now) {
      phase = 'expired';
      share = null;
    }
  }, 1000);
  onDestroy(() => clearInterval(tick));

  const remaining = $derived.by(() => {
    if (!share) {
      return '';
    }
    const seconds = Math.max(0, Math.floor((Date.parse(share.expiresAt) - now) / 1000));
    return `${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, '0')}`;
  });

  // The QR code as SVG markup, drawn from the hub's payload ("MT:…", in the
  // QR alphanumeric set) by the library.
  const qrSvg = $derived.by(() => {
    if (!share) {
      return '';
    }
    const qr = qrcode(0, 'M');
    qr.addData(share.qrCode, 'Alphanumeric');
    qr.make();
    return qr.createSvgTag({ cellSize: 4, margin: 2, scalable: true });
  });

  async function open() {
    phase = 'opening';
    problem = '';
    try {
      share = await shareNode(node);
      phase = 'open';
    } catch (err) {
      problem =
        err instanceof ApiError && err.code === 'busy'
          ? 'The device has a pairing window open already, perhaps from another app. Wait until it closes, then try again.'
          : err instanceof Error
            ? err.message
            : String(err);
      phase = 'confirm';
    }
  }

  async function close() {
    if (phase !== 'open') {
      onclose();
      return;
    }
    phase = 'closing';
    problem = '';
    try {
      await stopSharing(node);
      share = null;
      onclose();
    } catch (err) {
      problem = err instanceof Error ? err.message : String(err);
      phase = 'open';
    }
  }
</script>

<div class="share" aria-live="polite">
  {#if phase === 'confirm' || phase === 'opening'}
    <p>
      Let another app or hub — Google Home, Alexa, Apple Home, SmartThings — control <strong>{name}</strong> too.
      The hub opens a pairing window on the device for 10 minutes and shows a one-time setup code to add it with.
      The hub keeps controlling it.
    </p>
    {#if !online}<p class="problem">The device is offline; it must be online to open a pairing window.</p>{/if}
    {#if problem}<p class="problem">{problem}</p>{/if}
    <div class="actions">
      <button disabled={!online || phase === 'opening'} onclick={open}>
        {phase === 'opening' ? 'Opening…' : 'Open pairing window'}
      </button>
      <button class="quiet" disabled={phase === 'opening'} onclick={onclose}>Cancel</button>
    </div>
  {:else if phase === 'expired'}
    <p>The pairing window has closed. If the other app did not add the device, open a new one.</p>
    <div class="actions">
      <button onclick={open}>Open a new window</button>
      <button class="quiet" onclick={onclose}>Done</button>
    </div>
  {:else if share}
    <p>
      In the other app, add a Matter device and scan this code, or enter the setup code. The window closes in
      <strong>{remaining}</strong>.
    </p>
    <div class="codes">
      <div class="qr" role="img" aria-label="Matter QR code">{@html qrSvg}</div>
      <div>
        <p class="label">Setup code</p>
        <p class="manual">{share.manualCode}</p>
        <p class="hint">The code works once, while the window is open.</p>
      </div>
    </div>
    {#if problem}<p class="problem">{problem}</p>{/if}
    <div class="actions">
      <button class="quiet" disabled={phase === 'closing'} onclick={close}>
        {phase === 'closing' ? 'Closing…' : 'Close the window'}
      </button>
    </div>
  {/if}
</div>

<style>
  .share {
    border-top: 1px solid var(--border);
    padding-top: 0.75rem;
    margin-top: 0.75rem;
    display: grid;
    gap: 0.5rem;
  }

  .share p {
    margin: 0;
  }

  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .codes {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 1rem;
  }

  .qr {
    width: 168px;
    height: 168px;
    padding: 6px;
    background: #fff;
    border-radius: 8px;
    flex: none;
  }

  .qr :global(svg) {
    width: 100%;
    height: 100%;
    display: block;
  }

  .label {
    color: var(--muted);
    font-size: 0.8rem;
  }

  .manual {
    font-family: var(--mono, ui-monospace, monospace);
    font-size: 1.4rem;
    font-weight: 600;
    letter-spacing: 0.04em;
  }

  .hint {
    color: var(--muted);
    font-size: 0.85rem;
  }

  .problem {
    color: var(--danger);
  }
</style>
