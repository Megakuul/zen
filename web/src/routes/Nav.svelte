<script>
  import { onDestroy, onMount } from 'svelte';
  import { fade } from 'svelte/transition';

  let collapsed = $state(true);
  let hidden = $state(true);

  const keyHandler = (/** @type {KeyboardEvent}*/ e) => {
    if (e.key === 'Escape') collapsed = true;
  };

  let scrollTop = 0;
  const scrollHandler = (/** @type {Event}*/ e) => {
    if (collapsed) {
      if (scrollTop < window.scrollY) {
        hidden = false;
      } else {
        hidden = true;
      }
      scrollTop = window.scrollY;
    } else {
      e.stopPropagation(); // prevent scrolling in menu
    }
  };

  $effect(() => {
    if (typeof document === 'undefined') return;
    if (!collapsed) {
      document.addEventListener('keydown', keyHandler);
      return () => document.removeEventListener('keydown', keyHandler);
    }
  });

  onMount(() => {
    window.addEventListener('scroll', scrollHandler);
  });

  onDestroy(() => {
    if (typeof window === 'undefined') return;
    window.removeEventListener('scroll', scrollHandler);
  });
</script>

{#if collapsed}
  <nav
    transition:fade
    class="fixed w-full flex flex-row justify-center {hidden
      ? 'bottom-[-55px]'
      : 'bottom-10'} transition-all duration-700"
  >
    <button
      class="flex flex-row gap-4 justify-center items-center p-4 mx-2 w-96 text-2xl font-bold rounded-2xl cursor-pointer glass"
      onclick={() => {
        if (hidden) hidden = false;
        else collapsed = false;
      }}
    >
      <span class="text-3xl">Menu</span>
      <!-- prettier-ignore -->
      <svg class="w-8 h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><g fill="none"><path fill="url(#SVGZKAW8dxe)" d="M9.494 2L4.75 10.063l-.707 3.327a2.5 2.5 0 0 0 1.568.61l.017-.002V14h5.27a1.5 1.5 0 0 0 1.3-.75l2.597-4.5a1.5 1.5 0 0 0 0-1.5l-2.31-4A2.5 2.5 0 0 0 10.389 2l-.005.001V2z"/><path fill="url(#SVGGAy2DcWS)" fill-opacity="0.5" d="M9.494 2L4.75 10.063l-.707 3.327a2.5 2.5 0 0 0 1.568.61l.017-.002V14h5.27a1.5 1.5 0 0 0 1.3-.75l2.597-4.5a1.5 1.5 0 0 0 0-1.5l-2.31-4A2.5 2.5 0 0 0 10.389 2l-.005.001V2z"/><path fill="url(#SVGwpBLJ5HQ)" d="m9.051 3.11l.004-.016a1.5 1.5 0 0 1 1.367-1.092l-.034-.001h-.004L10.383 2H5.102a1.5 1.5 0 0 0-1.3.75l-2.597 4.5a1.5 1.5 0 0 0 0 1.5l2.31 4a2.5 2.5 0 0 0 1.388 1.126a1.5 1.5 0 0 0 2.055-1.02h.003l2.058-9.597q.013-.075.032-.149"/><path fill="url(#SVGQYqvHd4W)" fill-opacity="0.4" d="m9.051 3.11l.004-.016a1.5 1.5 0 0 1 1.367-1.092l-.034-.001h-.004L10.383 2H5.102a1.5 1.5 0 0 0-1.3.75l-2.597 4.5a1.5 1.5 0 0 0 0 1.5l2.31 4a2.5 2.5 0 0 0 1.388 1.126a1.5 1.5 0 0 0 2.055-1.02h.003l2.058-9.597q.013-.075.032-.149"/><defs><radialGradient id="SVGZKAW8dxe" cx="0" cy="0" r="1" gradientTransform="matrix(.77103 -17.4857 16.69068 .73597 8.924 17.257)" gradientUnits="userSpaceOnUse"><stop stop-color="#ffc470"/><stop offset=".251" stop-color="#ff835c"/><stop offset=".584" stop-color="#f24a9d"/><stop offset=".871" stop-color="#b339f0"/><stop offset="1" stop-color="#c354ff"/></radialGradient><radialGradient id="SVGGAy2DcWS" cx="0" cy="0" r="1" gradientTransform="matrix(-8.9317 -7.37295 7.32718 -8.87626 8.537 12.615)" gradientUnits="userSpaceOnUse"><stop offset=".709" stop-color="#ffb357" stop-opacity="0"/><stop offset=".942" stop-color="#ffb357"/></radialGradient><radialGradient id="SVGwpBLJ5HQ" cx="0" cy="0" r="1" gradientTransform="matrix(-16.24252 -5.82846 4.92524 -13.72546 13.488 12.8)" gradientUnits="userSpaceOnUse"><stop offset=".222" stop-color="#4e46e2"/><stop offset=".578" stop-color="#625df6"/><stop offset=".955" stop-color="#e37dff"/></radialGradient><linearGradient id="SVGQYqvHd4W" x1="3.875" x2="7.95" y1="6.971" y2="7.936" gradientUnits="userSpaceOnUse"><stop stop-color="#7563f7" stop-opacity="0"/><stop offset=".986" stop-color="#4916ae"/></linearGradient></defs></g></svg>
    </button>
  </nav>
{:else}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    transition:fade
    onclick={() => (collapsed = true)}
    class="flex fixed inset-0 z-40 flex-row justify-center w-screen h-dvh bg-slate-950/10 backdrop-blur-xl"
  ></div>
  <nav
    transition:fade
    class="z-50 fixed top-1/2 left-1/2 translate-y-[-50%] translate-x-[-50%] glass w-11/12 sm:w-10/12 lg:w-1/2 max-w-[1600px] p-10 rounded-2xl flex flex-col gap-4 text-2xl sm:text-4xl lg:text-6xl"
  >
    <a
      href="/"
      onclick={() => (collapsed = true)}
      class="flex flex-row gap-6 justify-start items-center p-3 rounded-xl transition-all duration-700 hover:scale-105 hover:glass hover:bg-slate-100/5"
    >
      <!-- prettier-ignore -->
      <svg class="w-14 h-14 sm:w-20 sm:h-20" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><g fill="none"><path fill="url(#SVGxRJaTbca)" d="M3 3.5A1.5 1.5 0 0 1 4.5 2h4A1.5 1.5 0 0 1 10 3.5V14H3.5a.5.5 0 0 1-.5-.5z"/><path fill="url(#SVG8ValwdZN)" d="M3 3.5A1.5 1.5 0 0 1 4.5 2h4A1.5 1.5 0 0 1 10 3.5V14H3.5a.5.5 0 0 1-.5-.5z"/><path fill="url(#SVG6SHu6dKT)" d="M3 3.5A1.5 1.5 0 0 1 4.5 2h4A1.5 1.5 0 0 1 10 3.5V14H3.5a.5.5 0 0 1-.5-.5z"/><path fill="url(#SVG7xoF8dDx)" d="M6 9.5a.5.5 0 1 1-1 0a.5.5 0 0 1 1 0"/><path fill="url(#SVG7xoF8dDx)" d="M8 7a.5.5 0 1 1-1 0a.5.5 0 0 1 1 0"/><path fill="url(#SVG7xoF8dDx)" d="M6 7a.5.5 0 1 1-1 0a.5.5 0 0 1 1 0"/><path fill="url(#SVG7xoF8dDx)" d="M8 4.5a.5.5 0 1 1-1 0a.5.5 0 0 1 1 0"/><path fill="url(#SVG7xoF8dDx)" d="M6 4.5a.5.5 0 1 1-1 0a.5.5 0 0 1 1 0"/><path fill="url(#SVGLOX9lx1q)" d="M10 11h3v4h-3z"/><path fill="url(#SVG24gbmdzd)" d="M8 11.46c0-.292.127-.569.349-.759l2.826-2.422a.5.5 0 0 1 .651 0l2.825 2.422c.221.19.349.467.349.76v3.04a.5.5 0 0 1-.5.5h-2v-2.5a.5.5 0 0 0-.5-.5h-1a.5.5 0 0 0-.5.5V15h-2a.5.5 0 0 1-.5-.5z"/><path fill="url(#SVGP6tngdLe)" fill-rule="evenodd" d="M10.518 7.36a1.5 1.5 0 0 1 1.964 0l3.26 2.823a.75.75 0 0 1-.983 1.134L11.5 8.493l-3.259 2.824a.75.75 0 0 1-.982-1.134z" clip-rule="evenodd"/><defs><linearGradient id="SVGxRJaTbca" x1="3" x2="13.981" y1="2.375" y2="10.576" gradientUnits="userSpaceOnUse"><stop stop-color="#29c3ff"/><stop offset="1" stop-color="#2764e7"/></linearGradient><linearGradient id="SVG7xoF8dDx" x1="5.9" x2="9.508" y1="3.333" y2="9.829" gradientUnits="userSpaceOnUse"><stop stop-color="#fdfdfd"/><stop offset="1" stop-color="#b3e0ff"/></linearGradient><linearGradient id="SVGLOX9lx1q" x1="11.5" x2="8.853" y1="11" y2="15.412" gradientUnits="userSpaceOnUse"><stop stop-color="#944600"/><stop offset="1" stop-color="#cd8e02"/></linearGradient><linearGradient id="SVG24gbmdzd" x1="7.764" x2="14.118" y1="8.349" y2="14.865" gradientUnits="userSpaceOnUse"><stop stop-color="#ffd394"/><stop offset="1" stop-color="#ffb357"/></linearGradient><linearGradient id="SVGP6tngdLe" x1="11.929" x2="11.193" y1="5.711" y2="11.112" gradientUnits="userSpaceOnUse"><stop stop-color="#ff921f"/><stop offset="1" stop-color="#eb4824"/></linearGradient><radialGradient id="SVG8ValwdZN" cx="0" cy="0" r="1" gradientTransform="matrix(0 4 -2.05556 0 9 13)" gradientUnits="userSpaceOnUse"><stop stop-color="#4a43cb"/><stop offset=".914" stop-color="#4a43cb" stop-opacity="0"/></radialGradient><radialGradient id="SVG6SHu6dKT" cx="0" cy="0" r="1" gradientTransform="matrix(-1.91666 2 -1.87874 -1.80045 9.417 11)" gradientUnits="userSpaceOnUse"><stop stop-color="#4a43cb"/><stop offset=".914" stop-color="#4a43cb" stop-opacity="0"/></radialGradient></defs></g></svg>
      <span>Home</span>
    </a>
    <a
      href="/planner"
      onclick={() => (collapsed = true)}
      class="flex flex-row gap-6 justify-start items-center p-3 rounded-xl transition-all duration-700 hover:scale-105 hover:bg-slate-100/5"
    >
      <!-- prettier-ignore -->
      <svg class="w-14 h-14 sm:w-20 sm:h-20" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><g fill="none"><path fill="url(#SVG8MbJrdML)" d="m11 14l3-3V5.5l-6-1l-6 1v6A2.5 2.5 0 0 0 4.5 14z"/><path fill="url(#SVGNDdPseIb)" d="m11 14l3-3V5.5l-6-1l-6 1v6A2.5 2.5 0 0 0 4.5 14z"/><path fill="url(#SVGZiw5lb0C)" fill-opacity="0.3" d="m11 14l3-3V5.5l-6-1l-6 1v6A2.5 2.5 0 0 0 4.5 14z"/><g filter="url(#SVGfoVamcHL)"><path fill="url(#SVGObTQ5KBI)" d="M5.248 8.997a.748.748 0 1 0 0-1.497a.748.748 0 0 0 0 1.497m.749 1.752a.748.748 0 1 1-1.497 0a.748.748 0 0 1 1.497 0M8 8.997A.748.748 0 1 0 8 7.5a.748.748 0 0 0 0 1.497m.749 1.752a.748.748 0 1 1-1.497 0a.748.748 0 0 1 1.497 0m2-1.752a.748.748 0 1 0 0-1.497a.748.748 0 0 0 0 1.497"/></g><path fill="url(#SVGCISDsA3i)" d="M14 4.5A2.5 2.5 0 0 0 11.5 2h-7A2.5 2.5 0 0 0 2 4.5V6h12z"/><path fill="url(#SVGrkdyidVB)" d="M11.352 9h2.64v2.646l-3.371 3.376a2.2 2.2 0 0 1-1.02.578l-1.496.375a.89.89 0 0 1-1.078-1.079l.374-1.498a2.2 2.2 0 0 1 .578-1.021z"/><path fill="url(#SVG5INsvlHL)" d="M10.485 15.143a2.2 2.2 0 0 1-.884.453l-1.496.375a.89.89 0 0 1-1.078-1.079l.374-1.498c.08-.318.229-.613.436-.864a3.5 3.5 0 0 0 2.648 2.613"/><path fill="url(#SVG2oyRacbK)" d="m11.54 8.82l1.271-1.272a1.87 1.87 0 0 1 2.642 2.644l-1.174 1.175z"/><path fill="url(#SVGGqqKUcjH)" d="M15.002 10.647A3.5 3.5 0 0 1 12.344 8l-.999 1a3.5 3.5 0 0 0 2.658 2.647z"/><defs><linearGradient id="SVG8MbJrdML" x1="10.167" x2="6.667" y1="15.167" y2="5" gradientUnits="userSpaceOnUse"><stop stop-color="#b3e0ff"/><stop offset="1" stop-color="#b3e0ff"/></linearGradient><linearGradient id="SVGNDdPseIb" x1="9.286" x2="11.025" y1="8.386" y2="16.154" gradientUnits="userSpaceOnUse"><stop stop-color="#dcf8ff" stop-opacity="0"/><stop offset="1" stop-color="#ff6ce8" stop-opacity="0.7"/></linearGradient><linearGradient id="SVGObTQ5KBI" x1="7.362" x2="8.566" y1="7.039" y2="15.043" gradientUnits="userSpaceOnUse"><stop stop-color="#0078d4"/><stop offset="1" stop-color="#0067bf"/></linearGradient><linearGradient id="SVGCISDsA3i" x1="2" x2="12.552" y1="2" y2="-.839" gradientUnits="userSpaceOnUse"><stop stop-color="#0094f0"/><stop offset="1" stop-color="#2764e7"/></linearGradient><linearGradient id="SVGrkdyidVB" x1="8.855" x2="12.286" y1="10.718" y2="14.149" gradientUnits="userSpaceOnUse"><stop stop-color="#ffa43d"/><stop offset="1" stop-color="#fb5937"/></linearGradient><linearGradient id="SVG5INsvlHL" x1="6.501" x2="9.001" y1="13.496" y2="15.993" gradientUnits="userSpaceOnUse"><stop offset=".255" stop-color="#ffd394"/><stop offset="1" stop-color="#ff921f"/></linearGradient><linearGradient id="SVG2oyRacbK" x1="15.067" x2="13.455" y1="7.909" y2="9.456" gradientUnits="userSpaceOnUse"><stop stop-color="#f97dbd"/><stop offset="1" stop-color="#dd3ce2"/></linearGradient><linearGradient id="SVGGqqKUcjH" x1="13.236" x2="10.655" y1="10.496" y2="9.364" gradientUnits="userSpaceOnUse"><stop stop-color="#ff921f"/><stop offset="1" stop-color="#ffe994"/></linearGradient><radialGradient id="SVGZiw5lb0C" cx="0" cy="0" r="1" gradientTransform="matrix(-5 5 -2.3139 -2.3139 11.5 12)" gradientUnits="userSpaceOnUse"><stop offset=".535" stop-color="#4a43cb"/><stop offset="1" stop-color="#4a43cb" stop-opacity="0"/></radialGradient><filter id="SVGfoVamcHL" width="9.664" height="6.664" x="3.167" y="6.833" color-interpolation-filters="sRGB" filterUnits="userSpaceOnUse"><feFlood flood-opacity="0" result="BackgroundImageFix"/><feColorMatrix in="SourceAlpha" result="hardAlpha" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0"/><feOffset dy=".667"/><feGaussianBlur stdDeviation=".667"/><feColorMatrix values="0 0 0 0 0.1242 0 0 0 0 0.323337 0 0 0 0 0.7958 0 0 0 0.32 0"/><feBlend in2="BackgroundImageFix" result="effect1_dropShadow_72095_10141"/><feBlend in="SourceGraphic" in2="effect1_dropShadow_72095_10141" result="shape"/></filter></defs></g></svg>
      <span>Planner</span>
    </a>
    <a
      href="/timer"
      onclick={() => (collapsed = true)}
      class="flex flex-row gap-6 justify-start items-center p-3 rounded-xl transition-all duration-700 hover:scale-105 hover:bg-slate-100/5"
    >
      <!-- prettier-ignore -->
      <svg class="w-14 h-14 sm:w-20 sm:h-20" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20"><g fill="none"><path fill="url(#SVGo0nuDeGv)" d="M10 18a8 8 0 1 0 0-16a8 8 0 0 0 0 16"/><path fill="url(#SVGvar9fccZ)" d="M10 5.25a4.75 4.75 0 0 0-3.455 8.01a.75.75 0 1 1-1.09 1.03a6.25 6.25 0 0 1 6.818-10.113a.75.75 0 1 1-.546 1.397A4.7 4.7 0 0 0 10 5.25"/><path fill="url(#SVG4JlvkeOO)" d="M10 5.25a4.75 4.75 0 0 0-3.455 8.01a.75.75 0 1 1-1.09 1.03a6.25 6.25 0 0 1 6.818-10.113a.75.75 0 1 1-.546 1.397A4.7 4.7 0 0 0 10 5.25"/><path fill="url(#SVGw48see6V)" d="M10 5.25a4.75 4.75 0 0 0-3.455 8.01a.75.75 0 1 1-1.09 1.03a6.25 6.25 0 0 1 6.818-10.113a.75.75 0 1 1-.546 1.397A4.7 4.7 0 0 0 10 5.25"/><path fill="url(#SVGvar9fccZ)" d="M14.852 7.301a.75.75 0 0 1 .972.426c.275.706.426 1.473.426 2.273c0 1.66-.649 3.171-1.705 4.29a.75.75 0 0 1-1.09-1.03A4.73 4.73 0 0 0 14.75 10a4.7 4.7 0 0 0-.324-1.727a.75.75 0 0 1 .426-.972"/><path fill="url(#SVG4JlvkeOO)" d="M14.852 7.301a.75.75 0 0 1 .972.426c.275.706.426 1.473.426 2.273c0 1.66-.649 3.171-1.705 4.29a.75.75 0 0 1-1.09-1.03A4.73 4.73 0 0 0 14.75 10a4.7 4.7 0 0 0-.324-1.727a.75.75 0 0 1 .426-.972"/><path fill="url(#SVGw48see6V)" d="M14.852 7.301a.75.75 0 0 1 .972.426c.275.706.426 1.473.426 2.273c0 1.66-.649 3.171-1.705 4.29a.75.75 0 0 1-1.09-1.03A4.73 4.73 0 0 0 14.75 10a4.7 4.7 0 0 0-.324-1.727a.75.75 0 0 1 .426-.972"/><path fill="url(#SVGAl4hJdJl)" d="M14.085 5.82a.5.5 0 0 1 .111.625l-.11.196l-.296.52l-.39.686l-.23.402l-.298.518a184 184 0 0 1-.99 1.69a30 30 0 0 1-.49.793l-.075.108a1.5 1.5 0 1 1-2.371-1.831c.072-.084.203-.204.343-.328c.15-.132.343-.296.56-.478c.436-.365.982-.81 1.515-1.242c.532-.43 1.054-.849 1.442-1.159l.274-.218l.37-.294a.5.5 0 0 1 .635.011"/><defs><linearGradient id="SVGo0nuDeGv" x1="7.714" x2="13.844" y1="2" y2="17.137" gradientUnits="userSpaceOnUse"><stop stop-color="#f4f4f4"/><stop offset="1" stop-color="#cbcbcb"/></linearGradient><linearGradient id="SVG4JlvkeOO" x1="8" x2="5" y1="11.5" y2="14" gradientUnits="userSpaceOnUse"><stop stop-color="#42b870" stop-opacity="0"/><stop offset=".716" stop-color="#42b870"/></linearGradient><linearGradient id="SVGw48see6V" x1="13" x2="15" y1="12" y2="14" gradientUnits="userSpaceOnUse"><stop stop-color="#e82c41" stop-opacity="0"/><stop offset=".563" stop-color="#e82c41"/></linearGradient><linearGradient id="SVGAl4hJdJl" x1="8.587" x2="12.566" y1="5.699" y2="11.081" gradientUnits="userSpaceOnUse"><stop stop-color="#0fafff"/><stop offset="1" stop-color="#2764e7"/></linearGradient><radialGradient id="SVGvar9fccZ" cx="0" cy="0" r="1" gradientTransform="matrix(13.5 0 0 20.2556 3.5 14.5)" gradientUnits="userSpaceOnUse"><stop stop-color="#42b870"/><stop offset=".59" stop-color="#ff921f"/><stop offset="1" stop-color="#e82c41"/></radialGradient></defs></g></svg>
      <span>Timer</span>
    </a>
    <a
      href="/leaderboard"
      onclick={() => (collapsed = true)}
      class="flex flex-row gap-6 justify-start items-center p-3 rounded-xl transition-all duration-700 hover:scale-105 hover:bg-slate-100/5"
    >
      <!-- prettier-ignore -->
      <svg class="w-14 h-14 sm:w-20 sm:h-20" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 28 28"><g fill="none"><path fill="url(#SVGUOn9CGiz)" d="M25 21.75A3.25 3.25 0 0 1 21.75 25H6.25A3.25 3.25 0 0 1 3 21.75V9l11-1l11 1z"/><path fill="url(#SVGMmlND7tz)" d="M25 21.75A3.25 3.25 0 0 1 21.75 25H6.25A3.25 3.25 0 0 1 3 21.75V9l11-1l11 1z"/><path fill="url(#SVGqFpwLenV)" fill-opacity="0.3" d="M25 21.75A3.25 3.25 0 0 1 21.75 25H6.25A3.25 3.25 0 0 1 3 21.75V9l11-1l11 1z"/><path fill="url(#SVG7HYjce0U)" fill-opacity="0.3" d="M25 21.75A3.25 3.25 0 0 1 21.75 25H6.25A3.25 3.25 0 0 1 3 21.75V9l11-1l11 1z"/><path fill="url(#SVGXxo55cJi)" fill-opacity="0.3" d="M25 21.75A3.25 3.25 0 0 1 21.75 25H6.25A3.25 3.25 0 0 1 3 21.75V9l11-1l11 1z"/><path fill="url(#SVGvJ5iCPCs)" d="M21.75 3A3.25 3.25 0 0 1 25 6.25V9H3V6.25A3.25 3.25 0 0 1 6.25 3z"/><path fill="url(#SVG6L6SOcHI)" d="M23 18.5a1.5 1.5 0 0 1 3 0v7a1.5 1.5 0 0 1-3 0z"/><path fill="url(#SVGem9bWd4h)" d="M20.5 14a1.5 1.5 0 0 0-1.5 1.5v10a1.5 1.5 0 0 0 3 0v-10a1.5 1.5 0 0 0-1.5-1.5"/><path fill="url(#SVGfzFX8bcw)" d="M16.5 20a1.5 1.5 0 0 0-1.5 1.5v4a1.5 1.5 0 0 0 3 0v-4a1.5 1.5 0 0 0-1.5-1.5"/><defs><linearGradient id="SVGUOn9CGiz" x1="17.972" x2="11.828" y1="27.088" y2="8.803" gradientUnits="userSpaceOnUse"><stop stop-color="#b3e0ff"/><stop offset="1" stop-color="#b3e0ff"/></linearGradient><linearGradient id="SVGMmlND7tz" x1="16.357" x2="19.402" y1="14.954" y2="28.885" gradientUnits="userSpaceOnUse"><stop stop-color="#dcf8ff" stop-opacity="0"/><stop offset="1" stop-color="#ff6ce8" stop-opacity="0.7"/></linearGradient><linearGradient id="SVGvJ5iCPCs" x1="3" x2="21.722" y1="3" y2="-3.157" gradientUnits="userSpaceOnUse"><stop stop-color="#0094f0"/><stop offset="1" stop-color="#2764e7"/></linearGradient><linearGradient id="SVG6L6SOcHI" x1="25.75" x2="24.291" y1="25.167" y2="16.87" gradientUnits="userSpaceOnUse"><stop stop-color="#d7257d"/><stop offset="1" stop-color="#e656eb"/></linearGradient><linearGradient id="SVGem9bWd4h" x1="22.75" x2="20.466" y1="28.444" y2="14.101" gradientUnits="userSpaceOnUse"><stop stop-color="#5b2ab5"/><stop offset="1" stop-color="#dd3ce2"/></linearGradient><linearGradient id="SVGfzFX8bcw" x1="15.375" x2="21.534" y1="20.292" y2="23.414" gradientUnits="userSpaceOnUse"><stop stop-color="#16bbda"/><stop offset="1" stop-color="#2052cb"/></linearGradient><radialGradient id="SVGqFpwLenV" cx="0" cy="0" r="1" gradientTransform="matrix(0 5 -2.54202 0 16.5 25)" gradientUnits="userSpaceOnUse"><stop offset=".535" stop-color="#4a43cb"/><stop offset="1" stop-color="#4a43cb" stop-opacity="0"/></radialGradient><radialGradient id="SVG7HYjce0U" cx="0" cy="0" r="1" gradientTransform="matrix(0 9.5 -2.5 0 20.5 22.5)" gradientUnits="userSpaceOnUse"><stop offset=".535" stop-color="#4a43cb"/><stop offset="1" stop-color="#4a43cb" stop-opacity="0"/></radialGradient><radialGradient id="SVGXxo55cJi" cx="0" cy="0" r="1" gradientTransform="rotate(90 .5 24)scale(6.5 2.5)" gradientUnits="userSpaceOnUse"><stop offset=".535" stop-color="#4a43cb"/><stop offset="1" stop-color="#4a43cb" stop-opacity="0"/></radialGradient></defs></g></svg>
      <span>Board</span>
    </a>
    <a
      href="/profile"
      onclick={() => (collapsed = true)}
      class="flex flex-row gap-6 justify-start items-center p-3 rounded-xl transition-all duration-700 hover:scale-105 hover:bg-slate-100/5"
    >
      <!-- prettier-ignore -->
      <svg class="w-14 h-14 sm:w-20 sm:h-20" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20"><g fill="none"><path fill="url(#SVGqRnNJbiH)" d="M2 5.75C2 4.784 2.784 4 3.75 4h12.5c.966 0 1.75.784 1.75 1.75v8.5A1.75 1.75 0 0 1 16.25 16H3.75A1.75 1.75 0 0 1 2 14.25z"/><path fill="url(#SVGYv2OvbWW)" fill-opacity="0.7" d="M2 5.75C2 4.784 2.784 4 3.75 4h12.5c.966 0 1.75.784 1.75 1.75v8.5A1.75 1.75 0 0 1 16.25 16H3.75A1.75 1.75 0 0 1 2 14.25z"/><path fill="url(#SVGiuqfcbae)" d="M11.5 8a.5.5 0 0 0 0 1h3a.5.5 0 0 0 0-1zm0 3a.5.5 0 0 0 0 1h3a.5.5 0 0 0 0-1z"/><path fill="url(#SVG8hZeWuhV)" d="M4 11.699a.95.95 0 0 1 .949-.949H8.05a.95.95 0 0 1 .949.949c0 .847-.577 1.585-1.399 1.791l-.059.015c-.684.17-1.4.17-2.084 0l-.06-.015A1.846 1.846 0 0 1 4 11.699"/><path fill="url(#SVGQys17bkF)" d="M8 8.5a1.5 1.5 0 1 1-3 0a1.5 1.5 0 0 1 3 0"/><defs><linearGradient id="SVGqRnNJbiH" x1="7.714" x2="11.389" y1="4" y2="16.099" gradientUnits="userSpaceOnUse"><stop stop-color="#b3e0ff"/><stop offset="1" stop-color="#8cd0ff"/></linearGradient><linearGradient id="SVGYv2OvbWW" x1="12.476" x2="15.575" y1="5.474" y2="22.667" gradientUnits="userSpaceOnUse"><stop offset=".447" stop-color="#ff6ce8" stop-opacity="0"/><stop offset="1" stop-color="#ff6ce8"/></linearGradient><linearGradient id="SVGiuqfcbae" x1="12.636" x2="14.653" y1="7.538" y2="15.199" gradientUnits="userSpaceOnUse"><stop stop-color="#0078d4"/><stop offset="1" stop-color="#0067bf"/></linearGradient><linearGradient id="SVG8hZeWuhV" x1="4" x2="5.105" y1="9" y2="14.055" gradientUnits="userSpaceOnUse"><stop offset=".125" stop-color="#9c6cfe"/><stop offset="1" stop-color="#7a41dc"/></linearGradient><linearGradient id="SVGQys17bkF" x1="5" x2="7.242" y1="6" y2="9.84" gradientUnits="userSpaceOnUse"><stop offset=".125" stop-color="#9c6cfe"/><stop offset="1" stop-color="#7a41dc"/></linearGradient></defs></g></svg>
      <span>Profile</span>
    </a>
  </nav>
{/if}
