<script>
  /**
   * @type {{
   *  value: string,
   *  increase?: number,
   *  factor?: number,
   *  list?: string,
   *  class?: string
   * }}
   */
  const {
    value,
    increase = 30,
    factor: baseFactor = 1,
    list = '012345679',
    class: className = undefined,
  } = $props();

  let displayValue = $derived.by(() => {
    let factor = baseFactor;
    let timeout = 0;
    for (let i = 0; i < value.length; i++) {
      for (let j = 0; j < list.length; j++) {
        timeout += increase + increase * factor;
        setTimeout(() => {
          displayValue = displayValue.slice(0, i) + list[j] + displayValue.slice(i + 1);
        }, timeout);
      }
      for (let j = 0; j < list.length; j++) {
        factor *= 1.01;
        if (list[j] === value[i]) break;
        timeout += increase + increase * factor;
        setTimeout(() => {
          displayValue = displayValue.slice(0, i) + list[j] + displayValue.slice(i + 1);
        }, timeout);
      }
      timeout += increase + increase * factor;
      setTimeout(() => {
        displayValue = displayValue.slice(0, i) + value[i] + displayValue.slice(i + 1);
      }, timeout);
      factor *= 2;
    }
    return '';
  });
</script>

<span class={className}>
  {displayValue}
</span>
