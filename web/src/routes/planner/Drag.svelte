<script lang="ts" generics="T">
  import { onMount } from 'svelte';

  interface Target {
    element: HTMLDivElement | undefined;
    data: T;
    grab: (x: number, y: number) => { x: number; y: number; width: number };
    move: (
      x: number,
      y: number,
      initX?: number,
      initY?: number,
    ) => { x: number; y: number; width: number };
    release: () => {};
  }

  interface Dropzone {
    element: HTMLDivElement | undefined;
    drop: (target: Target) => {};
  }

  let {
    targets,
    dropzones,
    draggedRef = $bindable(),
  }: {
    targets: Target[];
    dropzones: Dropzone[];
    draggedRef: T | undefined;
  } = $props();

  let maybeDrag = $state(0);
  let maybeDragY = $state(0);
  let initialDragX = $state(0);
  let initialDragY = $state(0);

  let dragged: Target | undefined = $state(undefined);

  $effect(() => {
    draggedRef = dragged?.data;
  });

  function handleDown(e: PointerEvent, target: Target) {
    maybeDragY = e.y;
    maybeDrag = setTimeout(() => {
      maybeDrag = 0;
      if (e.target instanceof Element) {
        e.target?.setPointerCapture(e.pointerId);
      }
      dragged = target;
      const { x, y, width } = target.grab(e.x, e.y);
      if (target.element)
        target.element.style = `position: fixed; width: ${width}px; left: ${x - width / 2}px; top: ${y}px;`;
      initialDragX = x - width / 2;
      initialDragY = y;
    }, 100);
  }

  function handleMove(e: PointerEvent) {
    if (Math.abs(maybeDragY - e.y) > 8) {
      clearTimeout(maybeDrag);
      maybeDrag = 0;
    }

    if (dragged && dragged.element) {
      e.preventDefault();
      e.stopPropagation();
      const { x, y, width } = dragged.move(e.x, e.y, initialDragX, initialDragY);
      dragged.element.style = `position: fixed; width: ${width}px; left: ${x - width / 2}px; top: ${y}px;`;
      initialDragX = x - width / 2;
      initialDragY = y;
    }
  }

  async function handleUp(e: PointerEvent) {
    clearTimeout(maybeDrag);
    maybeDrag = 0;

    e.preventDefault();
    e.stopPropagation();
    const target = dragged;
    dragged = undefined;
    if (!target) return;

    if (e.target instanceof Element && e.target.hasPointerCapture(e.pointerId)) {
      try {
        e.target?.releasePointerCapture(e.pointerId);
      } catch {
        /* ignore weird legacy browser specific failures */
      }
    }

    const underneath = document.elementFromPoint(e.x, e.y);
    for (const zone of dropzones) {
      if (underneath === target.element) {
        zone.drop(target);
        return;
      }
    }
    target.release();
  }

  onMount(() => {
    // the following interceptors are required because the browser dragging api is completely retarded.
    // first of all they must be implemented in js because they must not be "passive"
    // (which makes the browser bypass the eventbubbling in the decision phase (to get smoother scrolling)).
    // besides we want to allow scroll OR dragndrop so we cannot preventDefault ontouchstart (as we maybe need default scroll behavior).
    // therefore we just hijack and cancel operations that could be used by the browser to cancel the event (scroll or contextmenu).
    function interceptStart(e: TouchEvent) {
      if (dragged) e.preventDefault();
    }
    function interceptMove(e: TouchEvent) {
      if (maybeDrag || dragged) e.preventDefault();
    }
    function interceptContext(e: PointerEvent) {
      if (maybeDrag || dragged) e.preventDefault();
    }
    window.addEventListener('touchstart', interceptStart, { passive: false });
    window.addEventListener('touchmove', interceptMove, { passive: false });
    window.addEventListener('contextmenu', interceptContext, { passive: false });
    for (const target of targets) {
      target.element?.addEventListener('pointerdown', e => handleDown(e, target));
    }
    return () => {
      window.removeEventListener('touchstart', interceptStart);
      window.removeEventListener('touchmove', interceptMove);
      window.removeEventListener('contextmenu', interceptContext);
      for (const target of targets) {
        target.element?.removeEventListener('pointerdown', e => handleDown(e, target));
      }
    };
  });
</script>

<svelte:window onpointercancel={handleUp} onpointermove={handleMove} />
