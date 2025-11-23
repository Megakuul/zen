<!-- caution: extremly ugly code, please immediately forget what you saw here -->
<script>
  import { inview } from 'svelte-inview';
  import { Motion } from 'svelte-motion';

  /**
   * @type {{
   *  class?: string
   * }}
   */
  const { class: className = undefined } = $props();

  const delay = 0.05;
  let visible = $state(false);

  const calc = Object.entries({
    Sleep: { total: false, text: '441,504,000 seconds' },
    Work: { total: false, text: '441,504,000 seconds' },
    Meals: { total: false, text: '110,376,000 seconds' },
    Mother: { total: false, text: '55,188,000 seconds' },
    Budgerigar: { total: false, text: '13,797,000 seconds' },
    'Shopping, etc.': { total: false, text: '55,188,000 seconds' },
    'Friends, etc.': { total: false, text: '165,564,000 seconds' },
    'Miss Daria': { total: false, text: '27,594,000 seconds' },
    Daydreaming: { total: false, text: '13,797,000 seconds' },
    'Grand Total': { total: true, text: '1,324,512,000 seconds' },

    'Total time available': { total: false, text: '1,324,512,000 seconds' },
    'Time lost to date': { total: false, text: '1,324,512,000 seconds' },
    Balance: { total: true, text: '0.000,000,000 seconds' },
  });
</script>

<div class={className}>
  <div
    use:inview
    oninview_enter={() => (visible = true)}
    class="flex flex-col gap-4 justify-center items-center p-8 w-full text-xs rounded-2xl select-none sm:text-xl lg:w-11/12 lg:text-3xl chalk-font mirror-glass max-w-[1400px]"
  >
    {#each calc as [key, value], i}
      {@const [nextKey, nextValue] =
        calc[Math.max(Math.min(i + 1, calc.length - 1), calc.length - 1)]}
      {@const cumulativeDelay = i * (nextKey.length + nextValue.text.length) + 2}
      {#if value.total}
        <Motion
          variants={{
            hidden: { opacity: 0 },
            visible: i => ({
              opacity: 1,
              transition: { delay: (cumulativeDelay + nextKey.length) * delay },
            }),
          }}
          initial="hidden"
          animate={visible ? 'visible' : 'hidden'}
          let:motion
        >
          <hr use:motion style="opacity: 0;" class="w-full h-4 border-dotted" />
        </Motion>
      {/if}
      <span class="flex flex-row justify-start w-full">
        {#each (key + value.text).split('') as letter, j}
          <Motion
            variants={{
              hidden: { y: 50, opacity: 0 },
              visible: i => ({
                y: 0,
                opacity: 1,
                transition: { delay: (cumulativeDelay + j) * delay },
              }),
            }}
            initial="hidden"
            animate={visible ? 'visible' : 'hidden'}
            let:motion
          >
            <span
              use:motion
              style="opacity: 0;"
              class="{j === key.length ? 'ml-auto' : ''} {value.total ? 'font-bold' : ''}"
            >
              {#if letter === ' '}
                <span>&nbsp;</span>
              {:else}
                {letter}
              {/if}
            </span>
          </Motion>
        {/each}
      </span>
      {#if value.total}
        <br class="mb-10" />
      {/if}
    {/each}
  </div>

  <!-- agent -->
  <!-- prettier-ignore -->
  <svg class="w-11/12 sm:w-10/12 lg:w-1/2 translate-y-[-50px] sm:translate-y-[-160px]" viewBox="0 0 3139 1500" fill="none" xmlns="http://www.w3.org/2000/svg">
    <g class="agent-hand">
      <path d="M2295.37 686.073C2277.82 679.426 2258.87 691.747 2257.83 710.482L2246.1 921.417C2245.23 936.909 2256.33 950.502 2271.68 952.763L2276.21 953.43C2294.76 956.163 2311.25 941.392 2310.56 922.65L2307.01 826.088L2307 825.765L2307.02 825.442L2313.24 713.565C2313.92 701.484 2306.68 690.36 2295.37 686.073Z" fill="url(#paint0_radial_17_34)" stroke="black" stroke-width="14"/>
      <path d="M2098.68 708.71L2517.69 571.193C2525.04 568.782 2532.95 572.783 2535.36 580.13L2553.75 636.188C2555.65 641.96 2552.51 648.175 2546.73 650.07L2124.87 788.522C2102.83 795.755 2079.1 783.752 2071.87 761.713C2064.64 739.674 2076.64 715.943 2098.68 708.71Z" fill="url(#paint1_radial_17_34)" stroke="black" stroke-width="10"/>
      <g clip-path="url(#paint2_diamond_17_34_clip_path)" data-figma-skip-parse="true"><g transform="matrix(-0.014656 -0.0446565 0.0488589 -0.0160352 2261.35 699.508)"><rect x="0" y="0" width="1212.77" height="1194.47" fill="url(#paint2_diamond_17_34)" opacity="0.4" shape-rendering="crispEdges"/><rect x="0" y="0" width="1212.77" height="1194.47" transform="scale(1 -1)" fill="url(#paint2_diamond_17_34)" opacity="0.4" shape-rendering="crispEdges"/><rect x="0" y="0" width="1212.77" height="1194.47" transform="scale(-1 1)" fill="url(#paint2_diamond_17_34)" opacity="0.4" shape-rendering="crispEdges"/><rect x="0" y="0" width="1212.77" height="1194.47" transform="scale(-1)" fill="url(#paint2_diamond_17_34)" opacity="0.4" shape-rendering="crispEdges"/></g></g><path d="M2292.36 645.126L2318.55 724.938L2230.34 753.89L2204.14 674.078L2292.36 645.126Z" stroke="black" stroke-width="10"/>
      <path d="M2632.62 548.969C2633.24 545.089 2632.55 539.919 2631.63 535.547C2631.17 533.397 2630.68 531.504 2630.29 530.15C2630.1 529.473 2629.94 528.932 2629.83 528.563C2629.77 528.39 2629.73 528.254 2629.7 528.16C2629.66 528.085 2629.61 527.983 2629.54 527.856C2629.38 527.559 2629.15 527.126 2628.87 526.588C2628.29 525.511 2627.48 524.008 2626.55 522.308C2624.67 518.896 2622.31 514.725 2620.31 511.589C2612.73 499.698 2603.06 493.176 2593.46 486.121C2583.89 479.084 2574.49 471.562 2568.41 457.74C2563.67 446.968 2562.3 440.276 2560.2 428.889C2557.25 412.894 2558.93 392.304 2561.29 375.867C2561.51 374.38 2561.73 372.925 2561.95 371.508C2561.51 372.898 2561.06 374.327 2560.61 375.791C2555.78 391.72 2551.36 411.357 2552.62 426.879C2553.59 438.728 2554.47 445.411 2558.41 456.452C2563.61 470.984 2573.9 478.366 2585.04 485.925C2596.12 493.438 2608.09 501.143 2615.13 516.378C2620.16 527.274 2626.4 534.813 2625.05 548.001C2624.58 552.557 2622.57 557.024 2619.96 561.089C2617.34 565.168 2614.03 568.961 2610.82 572.187C2610.6 572.411 2610.38 572.629 2610.15 572.848C2611.29 572.2 2612.45 571.511 2613.62 570.781C2618.05 568.012 2622.5 564.701 2626 560.969C2629.51 557.227 2631.96 553.191 2632.62 548.969Z" fill="url(#paint3_linear_17_34)" stroke="black" stroke-width="4"/>
      <path d="M2572.84 559.882C2578.58 552.285 2582.64 549.006 2584.84 540.661C2586.44 534.566 2586.39 530.859 2585.53 524.451C2584.76 518.682 2583.82 515.498 2581.24 510.298C2575.75 499.195 2568.26 496.302 2557.4 488.327C2550.67 483.383 2541.19 478.426 2533.09 474.593C2533.83 475.178 2534.59 475.779 2535.36 476.397C2544.17 483.494 2554.85 492.773 2561.99 501.312C2565.37 505.356 2567.59 507.702 2570.53 512.387C2573.37 516.923 2575.54 519.823 2576.98 525.484C2578.79 532.597 2577.69 537.161 2576.71 543.822C2575.77 550.257 2572.77 557.559 2569.7 563.749C2570.84 562.418 2571.91 561.105 2572.84 559.882Z" fill="url(#paint4_linear_17_34)" stroke="black" stroke-width="4"/>
      <g filter="url(#filter0_f_17_34)">
      <path d="M2230.48 666.592C2231.65 645.535 2211.76 629.575 2191.45 635.281C2178.8 638.839 2169.8 650.049 2169.07 663.176L2164.29 749.155C2163.98 754.681 2164.92 760.205 2167.03 765.321L2202.73 851.874C2203.85 854.586 2204.64 857.421 2205.09 860.319L2212.36 907.219C2212.96 911.072 2214.16 914.807 2215.92 918.285L2235.19 956.298C2237.1 960.064 2238.35 964.13 2238.89 968.319L2253.27 1080.31C2256.01 1101.66 2276.32 1116.17 2297.4 1111.85L2363.48 1098.3C2366.66 1097.64 2369.73 1096.58 2372.63 1095.13L2384.1 1089.38C2400.8 1081.01 2408.65 1061.52 2402.4 1043.92L2397.65 1030.55C2396.43 1027.1 2394.7 1023.85 2392.52 1020.91L2336.16 944.793C2334.94 943.137 2333.85 941.382 2332.92 939.546L2231.41 739.89C2228.45 734.075 2227.08 727.58 2227.45 721.066L2230.48 666.592Z" fill="url(#paint5_radial_17_34)"/>
      <path d="M2192.67 639.614C2181.87 642.65 2174.19 652.22 2173.56 663.426L2168.78 749.405C2168.51 754.258 2169.33 759.11 2171.19 763.604L2206.89 850.158C2208.15 853.199 2209.04 856.38 2209.54 859.631L2216.81 906.53C2217.33 909.915 2218.39 913.195 2219.94 916.25L2239.21 954.263C2241.35 958.488 2242.75 963.048 2243.35 967.746L2257.73 1079.74C2260.14 1098.49 2277.98 1111.24 2296.49 1107.44L2362.58 1093.89C2365.37 1093.32 2368.07 1092.38 2370.61 1091.11L2382.09 1085.36C2396.75 1078.01 2403.64 1060.88 2398.16 1045.42L2393.41 1032.06C2392.33 1029.03 2390.82 1026.17 2388.9 1023.59L2332.55 947.47C2331.17 945.613 2329.96 943.644 2328.91 941.585L2227.4 741.93C2224.08 735.407 2222.55 728.123 2222.95 720.816L2225.98 666.343C2226.98 648.367 2210 634.742 2192.67 639.614Z" stroke="black" stroke-width="9"/>
      </g>
      <g filter="url(#filter1_f_17_34)">
      <path d="M2449.86 878.938C2452.5 857.924 2474.34 845.119 2493.97 853.076C2506.74 858.255 2514.81 870.984 2514.05 884.748L2509.32 969.661C2509.01 975.187 2507.47 980.573 2504.81 985.423L2459.72 1067.48C2458.31 1070.05 2457.21 1072.78 2456.44 1075.61L2444.01 1121.42C2442.99 1125.18 2441.38 1128.76 2439.25 1132.02L2412.38 1173L2382.01 1208.52C2376.5 1214.96 2368.96 1219.31 2360.63 1220.86L2333.26 1225.93C2321.21 1228.16 2308.84 1224.29 2300.22 1215.57L2287.53 1202.75C2285.95 1201.16 2284.53 1199.43 2283.26 1197.57L2278.37 1190.4C2271.01 1179.61 2269.88 1165.74 2275.41 1153.91L2302.49 1095.86C2304.37 1091.82 2306.97 1088.16 2310.15 1085.05L2350.88 1045.22L2433.82 953.235C2438.89 947.615 2442.11 940.577 2443.06 933.069L2449.86 878.938Z" fill="url(#paint6_radial_17_34)"/>
      <path d="M2492.27 857.246C2503.27 861.702 2510.21 872.655 2509.55 884.498L2504.83 969.411C2504.56 974.264 2503.2 978.995 2500.86 983.256L2455.78 1065.31C2454.19 1068.2 2452.96 1071.26 2452.09 1074.44L2439.67 1120.24C2438.77 1123.55 2437.36 1126.69 2435.48 1129.55L2408.77 1170.29L2378.59 1205.6C2373.75 1211.25 2367.12 1215.08 2359.81 1216.43L2332.44 1221.5C2321.86 1223.47 2310.99 1220.06 2303.42 1212.41L2290.73 1199.59C2289.34 1198.19 2288.09 1196.67 2286.98 1195.04L2282.09 1187.86C2275.62 1178.39 2274.63 1166.21 2279.48 1155.81L2306.57 1097.76C2308.22 1094.22 2310.5 1091 2313.3 1088.27L2354.03 1048.43L2354.13 1048.34L2354.22 1048.23L2437.17 956.248C2442.85 949.945 2446.47 942.051 2447.52 933.63L2454.32 879.498C2456.6 861.418 2475.39 850.399 2492.27 857.246Z" stroke="black" stroke-width="9"/>
      </g>
      <path d="M2452.91 1128.01C2482.12 1096.35 2204.28 1022.06 2222.49 1066.11C2240.7 1110.16 2105.26 1373.07 2105.26 1373.07L2361 1412.09C2361 1412.09 2423.7 1159.66 2452.91 1128.01Z" fill="url(#paint7_linear_17_34)"/>
      <path d="M2446.67 1118.26C2442.6 1114 2434.93 1109.02 2424.23 1103.72C2403.03 1093.21 2371.99 1082.45 2340.4 1073.98C2308.8 1065.52 2277.11 1059.48 2254.62 1058.24C2243.2 1057.61 2234.92 1058.28 2230.25 1059.99C2227.95 1060.84 2227.33 1061.64 2227.22 1061.85C2227.18 1061.92 2227.13 1062.03 2227.13 1062.28C2227.13 1062.57 2227.21 1063.13 2227.57 1064.01C2230.5 1071.1 2230.07 1081.42 2228.1 1092.88C2226.08 1104.66 2222.19 1118.83 2217.08 1134.37C2206.86 1165.47 2191.55 1202.66 2175.73 1238.34C2159.9 1274.04 2143.51 1308.35 2131.08 1333.72C2124.86 1346.41 2119.63 1356.88 2115.95 1364.17C2115.09 1365.89 2114.31 1367.43 2113.62 1368.78L2356.89 1405.9C2357.3 1404.26 2357.81 1402.22 2358.42 1399.83C2360.19 1392.85 2362.76 1382.85 2365.93 1370.76C2372.27 1346.57 2381.04 1313.98 2390.72 1280.39C2400.39 1246.82 2411 1212.18 2421.04 1183.91C2426.06 1169.78 2430.96 1157.16 2435.56 1147.03C2440.08 1137.08 2444.56 1128.94 2448.87 1124.28C2449.63 1123.44 2449.6 1123.02 2449.53 1122.62C2449.38 1121.86 2448.72 1120.39 2446.67 1118.26Z" stroke="url(#paint8_linear_17_34)" stroke-opacity="0.5" stroke-width="11"/>
    </g>

    <g filter="url(#filter2_f_17_34)">
    <g clip-path="url(#paint9_diamond_17_34_clip_path)" data-figma-skip-parse="true"><g transform="matrix(0.575046 0.492544 -0.468357 0.546808 1589.5 806.5)"><rect x="0" y="0" width="912.544" height="948.151" fill="url(#paint9_diamond_17_34)" opacity="1" shape-rendering="crispEdges"/><rect x="0" y="0" width="912.544" height="948.151" transform="scale(1 -1)" fill="url(#paint9_diamond_17_34)" opacity="1" shape-rendering="crispEdges"/><rect x="0" y="0" width="912.544" height="948.151" transform="scale(-1 1)" fill="url(#paint9_diamond_17_34)" opacity="1" shape-rendering="crispEdges"/><rect x="0" y="0" width="912.544" height="948.151" transform="scale(-1)" fill="url(#paint9_diamond_17_34)" opacity="1" shape-rendering="crispEdges"/></g></g><path d="M1986 611.5C1986 796.239 1801.77 1142.5 1574.5 1142.5C1347.23 1142.5 1163 796.239 1163 611.5C1163 426.761 1347.23 277 1574.5 277C1801.77 277 1986 426.761 1986 611.5Z" />
    <path d="M1574.5 285C1798.98 285 1978 432.648 1978 611.5C1978 701.811 1932.72 833.016 1858.86 941.783C1784.84 1050.77 1684.06 1134.5 1574.5 1134.5C1464.94 1134.5 1364.16 1050.77 1290.14 941.783C1216.28 833.016 1171 701.811 1171 611.5C1171 432.648 1350.02 285 1574.5 285Z" stroke="black" stroke-width="16"/>
    </g>
    
    <path d="M1879.78 996.671C1897.28 959.171 1709.86 1040.68 1549.28 1035.67C1400.01 1031.02 1207.28 940.67 1229.28 988.171C1251.28 1035.67 1296.28 1281.17 1290.78 1300.67C1285.28 1320.17 1533.28 1336.17 1533.28 1336.17L1775.28 1327.67C1816.09 1198.41 1862.28 1034.17 1879.78 996.671Z" fill="url(#paint10_linear_17_34)"/>
    <path d="M1875.31 992.954C1874.31 992.589 1872.48 992.28 1869.53 992.275C1866.66 992.269 1863.09 992.548 1858.82 993.103C1841.55 995.347 1815.13 1001.76 1782.56 1009.34C1717.88 1024.39 1630.22 1043.7 1549.11 1041.17C1473.7 1038.82 1387.52 1014.85 1323.63 997.674C1291.4 989.011 1265.2 982.173 1248.42 980.367C1244.27 979.921 1240.89 979.806 1238.26 980.002C1235.54 980.203 1234.09 980.703 1233.42 981.103C1233.02 981.342 1233.09 981.378 1233.07 981.497C1233.01 981.981 1233.09 983.309 1234.27 985.859C1240.04 998.309 1247.09 1023.09 1254.3 1053.01C1261.55 1083.12 1269.08 1118.96 1275.75 1153.89C1282.42 1188.82 1288.23 1222.91 1292.03 1249.52C1293.93 1262.81 1295.34 1274.3 1296.1 1283.12C1296.48 1287.52 1296.71 1291.33 1296.74 1294.41C1296.76 1296.48 1296.71 1298.53 1296.46 1300.28C1297.41 1300.9 1299.03 1301.71 1301.53 1302.63C1307.33 1304.77 1316.2 1306.93 1327.37 1309.04C1349.62 1313.24 1379.92 1317.05 1410.55 1320.26C1441.14 1323.48 1471.91 1326.08 1495.05 1327.88C1506.61 1328.78 1516.26 1329.48 1523.02 1329.96C1526.4 1330.2 1529.06 1330.38 1530.87 1330.5C1531.77 1330.56 1532.47 1330.61 1532.93 1330.64C1533.11 1330.65 1533.26 1330.66 1533.37 1330.66L1771.21 1322.31C1791.26 1258.63 1812.52 1186.94 1831.01 1126.16C1849.78 1064.48 1865.84 1013.55 1874.8 994.346C1875.09 993.721 1875.24 993.267 1875.33 992.959C1875.32 992.958 1875.32 992.956 1875.31 992.954Z" stroke="url(#paint11_linear_17_34)" stroke-opacity="0.5" stroke-width="11"/>
    <path d="M1569.5 287.5C1729.68 287.5 1874.31 305.68 1978.56 334.854C2030.76 349.461 2072.24 366.661 2100.43 385.31C2129.03 404.23 2141.5 422.926 2141.5 440C2141.5 457.074 2129.03 475.77 2100.43 494.69C2072.24 513.339 2030.76 530.539 1978.56 545.146C1874.31 574.32 1729.68 592.5 1569.5 592.5C1409.32 592.5 1264.69 574.32 1160.44 545.146C1108.24 530.539 1066.76 513.339 1038.57 494.69C1009.97 475.77 997.5 457.074 997.5 440C997.5 422.926 1009.97 404.23 1038.57 385.31C1066.76 366.661 1108.24 349.461 1160.44 334.854C1264.69 305.68 1409.32 287.5 1569.5 287.5Z" fill="url(#paint12_radial_17_34)" stroke="black" stroke-width="21"/>
    <path d="M1317.77 20.0459C1340.29 11.5475 1363.11 9.55293 1388.04 10.8711C1413.25 12.204 1439.81 16.8429 1470.43 21.3242C1500.79 25.7667 1534.43 29.9248 1572.5 29.9062C1610.37 29.8877 1644.57 25.7429 1675.84 21.3359C1707.35 16.8943 1735.3 12.2913 1761.87 10.9932C1788.21 9.7061 1812.42 11.737 1835.96 20.2686C1859.43 28.7807 1882.9 44.0101 1907.34 70.0859C1924.03 87.8933 1936.58 118.464 1945.59 156.08C1954.52 193.332 1959.64 235.961 1962.46 276.449C1965.28 316.885 1965.8 354.9 1965.6 382.825C1965.5 396.78 1965.23 408.195 1964.98 416.11C1964.93 417.781 1964.88 419.296 1964.83 420.646C1963.93 420.869 1962.97 421.112 1961.92 421.37C1955.14 423.052 1945.25 425.459 1932.9 428.353C1908.2 434.141 1873.64 441.875 1834.22 449.648C1755.21 465.232 1657.39 480.838 1580.41 481.5C1499.05 482.2 1395.52 466.598 1311.88 450.67C1270.16 442.724 1233.56 434.734 1207.4 428.73C1194.32 425.729 1183.85 423.225 1176.66 421.474C1175.56 421.204 1174.53 420.951 1173.58 420.719C1173.6 419.515 1173.62 418.188 1173.65 416.742C1173.8 408.89 1174.1 397.561 1174.69 383.702C1175.89 355.97 1178.28 318.182 1183.05 277.892C1187.83 237.555 1194.97 194.97 1205.59 157.554C1216.29 119.828 1230.17 88.7094 1247.65 70.0938C1272.34 43.8088 1295.27 28.5403 1317.77 20.0459Z" fill="url(#paint13_radial_17_34)" stroke="black" stroke-width="21"/>

    <defs>
      <clipPath id="paint2_diamond_17_34_clip_path"><path d="M2292.36 645.126L2318.55 724.938L2230.34 753.89L2204.14 674.078L2292.36 645.126Z" stroke-width="10"/></clipPath><filter id="filter0_f_17_34" x="2160.23" y="630.113" width="248.305" height="486.499" filterUnits="userSpaceOnUse" color-interpolation-filters="sRGB">
      <feFlood flood-opacity="0" result="BackgroundImageFix"/>
      <feBlend mode="normal" in="SourceGraphic" in2="BackgroundImageFix" result="shape"/>
      <feGaussianBlur stdDeviation="2" result="effect1_foregroundBlur_17_34"/>
      </filter>
      <filter id="filter1_f_17_34" x="2267.94" y="846.692" width="250.16" height="383.858" filterUnits="userSpaceOnUse" color-interpolation-filters="sRGB">
      <feFlood flood-opacity="0" result="BackgroundImageFix"/>
      <feBlend mode="normal" in="SourceGraphic" in2="BackgroundImageFix" result="shape"/>
      <feGaussianBlur stdDeviation="2" result="effect1_foregroundBlur_17_34"/>
      </filter>
      <filter id="filter2_f_17_34" x="1155" y="269" width="839" height="881.5" filterUnits="userSpaceOnUse" color-interpolation-filters="sRGB">
      <feFlood flood-opacity="0" result="BackgroundImageFix"/>
      <feBlend mode="normal" in="SourceGraphic" in2="BackgroundImageFix" result="shape"/>
      <feGaussianBlur stdDeviation="4" result="effect1_foregroundBlur_17_34"/>
      </filter>
      <clipPath id="paint9_diamond_17_34_clip_path"><path d="M1986 611.5C1986 796.239 1801.77 1142.5 1574.5 1142.5C1347.23 1142.5 1163 796.239 1163 611.5C1163 426.761 1347.23 277 1574.5 277C1801.77 277 1986 426.761 1986 611.5Z"/></clipPath><radialGradient id="paint0_radial_17_34" cx="0" cy="0" r="1" gradientUnits="userSpaceOnUse" gradientTransform="translate(2299.97 795.848) rotate(105.401) scale(168 38.5)">
      <stop stop-color="#161616"/>
      <stop offset="1"/>
      </radialGradient>
      <radialGradient id="paint1_radial_17_34" cx="0" cy="0" r="1" gradientUnits="userSpaceOnUse" gradientTransform="translate(2307.98 684.223) rotate(161.83) scale(253.5 47)">
      <stop stop-color="#242424"/>
      <stop offset="1" stop-color="#2F2F2F"/>
      </radialGradient>
      <linearGradient id="paint2_diamond_17_34" x1="0" y1="0" x2="500" y2="500" gradientUnits="userSpaceOnUse">
      <stop stop-color="#684802"/>
      <stop offset="1" stop-color="#3C2901"/>
      </linearGradient>
      <linearGradient id="paint3_linear_17_34" x1="2554.83" y1="467.246" x2="2625.26" y2="459.889" gradientUnits="userSpaceOnUse">
      <stop stop-color="#5C5C5C"/>
      <stop offset="1" stop-color="#333333"/>
      </linearGradient>
      <linearGradient id="paint4_linear_17_34" x1="2501.09" y1="488.851" x2="2582.78" y2="549.919" gradientUnits="userSpaceOnUse">
      <stop stop-color="#5C5C5C"/>
      <stop offset="1" stop-color="#333333"/>
      </linearGradient>
      <radialGradient id="paint5_radial_17_34" cx="0" cy="0" r="1" gradientUnits="userSpaceOnUse" gradientTransform="translate(2215.64 768.222) rotate(74.6414) scale(154.735 51.0226)">
      <stop/>
      <stop offset="1" stop-color="#111010"/>
      </radialGradient>
      <radialGradient id="paint6_radial_17_34" cx="0" cy="0" r="1" gradientUnits="userSpaceOnUse" gradientTransform="translate(2459.85 983.119) rotate(111.727) scale(154.735 51.0226)">
      <stop/>
      <stop offset="1" stop-color="#111010"/>
      </radialGradient>
      <linearGradient id="paint7_linear_17_34" x1="2297.5" y1="1346" x2="2347.37" y2="1095.63" gradientUnits="userSpaceOnUse">
      <stop stop-color="#282828" stop-opacity="0"/>
      <stop offset="0.485215" stop-color="#282828"/>
      <stop offset="1" stop-color="#262626"/>
      </linearGradient>
      <linearGradient id="paint8_linear_17_34" x1="2309.5" y1="1329" x2="2357" y2="1072.5" gradientUnits="userSpaceOnUse">
      <stop stop-opacity="0"/>
      <stop offset="1"/>
      </linearGradient>
      <linearGradient id="paint9_diamond_17_34" x1="0" y1="0" x2="500" y2="500" gradientUnits="userSpaceOnUse">
      <stop/>
      <stop offset="1" stop-color="#0C0C0C"/>
      </linearGradient>
      <linearGradient id="paint10_linear_17_34" x1="1577.5" y1="1031.5" x2="1560" y2="1268.5" gradientUnits="userSpaceOnUse">
      <stop stop-color="#141414"/>
      <stop offset="0.418424" stop-color="#1D1D1D"/>
      <stop offset="0.598678" stop-color="#212121" stop-opacity="0.730285"/>
      <stop offset="1" stop-color="#2B2B2B" stop-opacity="0"/>
      </linearGradient>
      <linearGradient id="paint11_linear_17_34" x1="1565.5" y1="1027.5" x2="1560" y2="1263" gradientUnits="userSpaceOnUse">
      <stop/>
      <stop offset="1" stop-opacity="0"/>
      </linearGradient>
      <radialGradient id="paint12_radial_17_34" cx="0" cy="0" r="1" gradientUnits="userSpaceOnUse" gradientTransform="translate(1569.5 440) rotate(90) scale(163 582.5)">
      <stop stop-color="#3D3D3D"/>
      <stop offset="1" stop-color="#1A1A1A"/>
      </radialGradient>
      <radialGradient id="paint13_radial_17_34" cx="0" cy="0" r="1" gradientUnits="userSpaceOnUse" gradientTransform="translate(1569.57 246.012) rotate(90) scale(246.012 406.569)">
      <stop stop-color="#3D3D3D"/>
      <stop offset="1" stop-color="#1A1A1A"/>
      </radialGradient>
    </defs>
  </svg>

  <div
    use:inview
    oninview_enter={(/** @type {any} */ e) => {
      e.target?.classList.add('pull-up');
    }}
    oninview_leave={(/** @type {any} */ e) => {
      e.target?.classList.remove('pull-up');
    }}
    class="p-4 w-11/12 text-xs translate-y-0 sm:p-8 sm:text-base lg:w-10/12 lg:text-2xl glass bubble max-w-[1200px] sm:translate-y-[-150px]"
    style="clip-path: url(#bubble-clip);"
  >
    It's like this, my dear sir, you're wasting your life cutting hair, lathering faces and swapping
    idle chitchat. When you're dead, it'll be as if you'd never existed. If you only had the time to
    lead the right kind of life, you'd be quite a different person. <a
      href="/getting-started"
      class="underline">Time is all you need, right?</a
    >
  </div>

  <p
    class="text-xs text-center sm:text-base text-slate-200/30 translate-y-[10px] sm:translate-y-[-80px]"
  >
    didn't get the joke? read <a class="underline" href="https://en.wikipedia.org/wiki/Momo_(novel)"
      >this</a
    >
  </p>
</div>

<style>
  .chalk-font {
    font-family: 'Cabin Sketch', sans-serif;
    font-weight: 400;
    font-style: normal;
  }

  .mirror-glass {
    box-shadow:
      rgba(255, 255, 255, 0.04) 0px 6px 24px 0px,
      rgba(255, 255, 255, 0.08) 0px 0px 0px 1px;
    background-color: rgba(255, 255, 255, 0.005);
    backdrop-filter: blur(2px);
  }

  .agent-hand {
    transform-origin: center center;
    animation: shake linear 6s infinite;
  }

  @keyframes shake {
    0%,
    100% {
      transform: rotateZ(0deg);
    }
    25% {
      transform: rotateZ(3deg);
    }
    75% {
      transform: rotateZ(-3deg);
    }
  }

  .bubble {
    transition: all ease 1s;
    font-family: 'Bebas Neue', sans-serif;
    font-weight: 400;
    font-style: normal;
  }

  .bubble:after {
    content: '';
    position: absolute;
    top: 0;
    left: 5%;
    width: 0;
    height: 0;
    border: 40px solid transparent;
    border-bottom-color: rgba(255, 255, 255, 0.08);
    border-top: 0;
    border-right: 0;
    margin-left: -19px;
    margin-top: -50px;
  }

  @media only screen and (max-width: 500px) {
    .bubble:after {
      left: 12%;
      border: 20px solid transparent;
      border-bottom-color: rgba(255, 255, 255, 0.08);
      border-top: 0;
      border-right: 0;
      margin-left: -20px;
      margin-top: -25px;
    }
  }

  .pull-up {
    transform: translateY(-20px);
  }
</style>

