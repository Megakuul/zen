<script>
  import { inview } from "svelte-inview";
  import { Motion } from "svelte-motion";

  /**
   * @type {{
   *  visible: boolean,
   *  delay?: number,
   * }}
   */
  let {
    visible = $bindable(),
    delay = 0.05,
  } = $props();

  const calc = Object.entries({
    "Sleep": {total: false, text: "441,504,000 seconds"},
    "Work": {total: false, text: "441,504,000 seconds"},
    "Meals": {total: false, text: "110,376,000 seconds"},
    "Mother": {total: false, text: "55,188,000 seconds"},
    "Budgerigar": {total: false, text: "13,797,000 seconds"},
    "Shopping, etc.": {total: false, text: "55,188,000 seconds"},
    "Friends, etc.": {total: false, text: "165,564,000 seconds"},
    "Miss Daria": {total: false, text: "27,594,000 seconds"},
    "Daydreaming": {total: false, text: "13,797,000 seconds"},
    "Grand Total": {total: true, text: "1,324,512,000 seconds"},

    "Total time available": {total: false, text: "1,324,512,000 seconds"},
    "Time lost to date": {total: false, text: "1,324,512,000 seconds"},
    "Balance": {total: true, text: "0.000,000,000 seconds"},
  })
</script>

<div use:inview oninview_enter={() => visible = true} class="chalk-font mirror-glass w-full flex flex-col justify-center items-center gap-4 p-8 rounded-2xl text-xs sm:text-xl lg:text-3xl select-none">
  {#each calc as [key, value], i}
    {@const [nextKey, nextValue] = calc[Math.max(Math.min(i+1, calc.length-1), calc.length-1)]}
    {@const cumulativeDelay = i * (nextKey.length + nextValue.text.length) + 2}
    {#if value.total}
      <Motion         
        variants={{
          hidden: { opacity: 0 },
          visible: (i) => ({ opacity: 1, transition: { delay: (cumulativeDelay + nextKey.length) * delay }})
        }}
        initial="hidden"
        animate={visible ? "visible" : "hidden"}
        let:motion>
        <hr use:motion class="w-full h-4 border-dotted">
      </Motion>
    {/if}
    <span class="flex flex-row justify-start w-full">
      {#each (key + value.text).split("") as letter, j}
      <Motion
        variants={{
          hidden: { y: 50, opacity: 0 },
          visible: (i) => ({ y: 0, opacity: 1, transition: { delay: (cumulativeDelay + j) * delay }})
        }}
        initial="hidden"
        animate={visible ? "visible" : "hidden"}
        let:motion>
        <span use:motion class="{j === key.length ? "ml-auto" : ""} {value.total ? "font-bold" : ""}">
          {#if letter === " "}
            <span>&nbsp;</span>
          {:else}
            {letter}
          {/if}
        </span>
      </Motion>
      {/each}
    </span>
    {#if value.total}
      <br class="mb-10">
    {/if}
  {/each}
</div>

<!-- gray man -->
<svg viewBox="0 0 1768 1526" fill="none" xmlns="http://www.w3.org/2000/svg">
  <g filter="url(#filter0_n_17_18)">
  <path d="M1308.37 686.073C1290.82 679.426 1271.87 691.747 1270.83 710.482L1259.1 921.417C1258.23 936.909 1269.33 950.502 1284.68 952.763L1289.21 953.43C1307.76 956.163 1324.25 941.392 1323.56 922.65L1320.01 826.088L1320 825.765L1320.02 825.442L1326.24 713.565C1326.92 701.484 1319.68 690.36 1308.37 686.073Z" fill="url(#paint0_radial_17_18)" stroke="black" stroke-width="14"/>
  <path d="M1111.68 708.71L1530.69 571.193C1538.04 568.782 1545.95 572.783 1548.36 580.13L1566.75 636.188C1568.65 641.96 1565.51 648.175 1559.73 650.07L1137.87 788.522C1115.83 795.755 1092.1 783.752 1084.87 761.713C1077.64 739.674 1089.64 715.943 1111.68 708.71Z" fill="url(#paint1_radial_17_18)" stroke="black" stroke-width="10"/>
  <g clip-path="url(#paint2_diamond_17_18_clip_path)" data-figma-skip-parse="true"><g transform="matrix(-0.014656 -0.0446565 0.0488589 -0.0160352 1274.35 699.508)"><rect x="0" y="0" width="1212.77" height="1194.47" fill="url(#paint2_diamond_17_18)" opacity="0.4" shape-rendering="crispEdges"/><rect x="0" y="0" width="1212.77" height="1194.47" transform="scale(1 -1)" fill="url(#paint2_diamond_17_18)" opacity="0.4" shape-rendering="crispEdges"/><rect x="0" y="0" width="1212.77" height="1194.47" transform="scale(-1 1)" fill="url(#paint2_diamond_17_18)" opacity="0.4" shape-rendering="crispEdges"/><rect x="0" y="0" width="1212.77" height="1194.47" transform="scale(-1)" fill="url(#paint2_diamond_17_18)" opacity="0.4" shape-rendering="crispEdges"/></g></g><path d="M1305.36 645.126L1331.55 724.938L1243.34 753.89L1217.14 674.078L1305.36 645.126Z" stroke="black" stroke-width="10"/>
  <path d="M1645.62 548.969C1646.24 545.089 1645.55 539.919 1644.63 535.547C1644.17 533.397 1643.68 531.504 1643.29 530.15C1643.1 529.473 1642.94 528.932 1642.83 528.563C1642.77 528.39 1642.73 528.254 1642.7 528.16C1642.66 528.085 1642.61 527.983 1642.54 527.856C1642.38 527.559 1642.15 527.126 1641.87 526.588C1641.29 525.511 1640.48 524.008 1639.55 522.308C1637.67 518.896 1635.31 514.725 1633.31 511.589C1625.73 499.698 1616.06 493.176 1606.46 486.121C1596.89 479.084 1587.49 471.562 1581.41 457.74C1576.67 446.968 1575.3 440.276 1573.2 428.889C1570.25 412.894 1571.93 392.304 1574.29 375.867C1574.51 374.38 1574.73 372.925 1574.95 371.508C1574.51 372.898 1574.06 374.327 1573.61 375.791C1568.78 391.72 1564.36 411.357 1565.62 426.879C1566.59 438.728 1567.47 445.411 1571.41 456.452C1576.61 470.984 1586.9 478.366 1598.04 485.925C1609.12 493.438 1621.09 501.143 1628.13 516.378C1633.16 527.274 1639.4 534.813 1638.05 548.001C1637.58 552.557 1635.57 557.024 1632.96 561.089C1630.34 565.168 1627.03 568.961 1623.82 572.187C1623.6 572.411 1623.38 572.629 1623.15 572.848C1624.29 572.2 1625.45 571.511 1626.62 570.781C1631.05 568.012 1635.5 564.701 1639 560.969C1642.51 557.227 1644.96 553.191 1645.62 548.969Z" fill="url(#paint3_linear_17_18)" stroke="black" stroke-width="4"/>
  <path d="M1585.84 559.882C1591.58 552.285 1595.64 549.006 1597.84 540.661C1599.44 534.566 1599.39 530.859 1598.53 524.451C1597.76 518.682 1596.82 515.498 1594.24 510.298C1588.75 499.195 1581.26 496.302 1570.4 488.327C1563.67 483.383 1554.19 478.426 1546.09 474.593C1546.83 475.178 1547.59 475.779 1548.36 476.397C1557.17 483.494 1567.85 492.773 1574.99 501.312C1578.37 505.356 1580.59 507.702 1583.53 512.387C1586.37 516.923 1588.54 519.823 1589.98 525.484C1591.79 532.597 1590.69 537.161 1589.71 543.822C1588.77 550.257 1585.77 557.559 1582.7 563.749C1583.84 562.418 1584.91 561.105 1585.84 559.882Z" fill="url(#paint4_linear_17_18)" stroke="black" stroke-width="4"/>
  <g filter="url(#filter1_f_17_18)">
  <path d="M1243.48 666.592C1244.65 645.535 1224.76 629.575 1204.45 635.281C1191.8 638.839 1182.8 650.049 1182.07 663.176L1177.29 749.155C1176.98 754.681 1177.92 760.205 1180.03 765.321L1215.73 851.874C1216.85 854.586 1217.64 857.421 1218.09 860.319L1225.36 907.219C1225.96 911.072 1227.16 914.807 1228.92 918.285L1248.19 956.298C1250.1 960.064 1251.35 964.13 1251.89 968.319L1266.27 1080.31C1269.01 1101.66 1289.32 1116.17 1310.4 1111.85L1376.48 1098.3C1379.66 1097.64 1382.73 1096.58 1385.63 1095.13L1397.1 1089.38C1413.8 1081.01 1421.65 1061.52 1415.4 1043.92L1410.65 1030.55C1409.43 1027.1 1407.7 1023.85 1405.52 1020.91L1349.16 944.793C1347.94 943.137 1346.85 941.382 1345.92 939.546L1244.41 739.89C1241.45 734.075 1240.08 727.58 1240.45 721.066L1243.48 666.592Z" fill="url(#paint5_radial_17_18)"/>
  <path d="M1205.67 639.614C1194.87 642.65 1187.19 652.22 1186.56 663.426L1181.78 749.405C1181.51 754.258 1182.33 759.11 1184.19 763.604L1219.89 850.158C1221.15 853.199 1222.04 856.38 1222.54 859.631L1229.81 906.53C1230.33 909.915 1231.39 913.195 1232.94 916.25L1252.21 954.263C1254.35 958.488 1255.75 963.048 1256.35 967.746L1270.73 1079.74C1273.14 1098.49 1290.98 1111.24 1309.49 1107.44L1375.58 1093.89C1378.37 1093.32 1381.07 1092.38 1383.61 1091.11L1395.09 1085.36C1409.75 1078.01 1416.64 1060.88 1411.16 1045.42L1406.41 1032.06C1405.33 1029.03 1403.82 1026.17 1401.9 1023.59L1345.55 947.47C1344.17 945.613 1342.96 943.644 1341.91 941.585L1240.4 741.93C1237.08 735.407 1235.55 728.123 1235.95 720.816L1238.98 666.343C1239.98 648.367 1223 634.742 1205.67 639.614Z" stroke="black" stroke-width="9"/>
  </g>
  <g filter="url(#filter2_f_17_18)">
  <path d="M1462.86 878.938C1465.5 857.924 1487.34 845.119 1506.97 853.076C1519.74 858.255 1527.81 870.984 1527.05 884.748L1522.32 969.661C1522.01 975.187 1520.47 980.573 1517.81 985.423L1472.72 1067.48C1471.31 1070.05 1470.21 1072.78 1469.44 1075.61L1457.01 1121.42C1455.99 1125.18 1454.38 1128.76 1452.25 1132.02L1428.88 1167.66C1426.56 1171.19 1424.87 1175.09 1423.87 1179.2L1397.16 1288.91C1392.07 1309.82 1370.27 1321.99 1349.8 1315.35L1308.34 1301.91C1292.69 1296.84 1283.23 1280.95 1286.23 1264.77L1295.04 1217.22C1296.32 1210.27 1289.13 1204.83 1282.8 1207.97C1275.61 1211.53 1267.91 1204.16 1271.14 1196.82L1315.45 1096.29C1317.36 1091.98 1320.07 1088.06 1323.44 1084.76L1363.88 1045.22L1446.82 953.235C1451.89 947.615 1455.11 940.577 1456.06 933.069L1462.86 878.938Z" fill="url(#paint6_radial_17_18)"/>
  <path d="M1505.27 857.245C1516.27 861.702 1523.21 872.655 1522.55 884.499L1517.83 969.411C1517.56 974.265 1516.2 978.995 1513.86 983.256L1468.78 1065.31C1467.19 1068.2 1465.96 1071.26 1465.09 1074.44L1452.67 1120.24C1451.77 1123.55 1450.36 1126.69 1448.48 1129.55L1425.11 1165.19C1422.52 1169.15 1420.62 1173.53 1419.5 1178.13L1392.79 1287.84C1388.32 1306.21 1369.17 1316.9 1351.19 1311.07L1309.73 1297.63C1296.23 1293.26 1288.06 1279.55 1290.65 1265.59L1299.46 1218.04C1301.42 1207.44 1290.45 1199.15 1280.8 1203.94C1277.38 1205.63 1273.72 1202.13 1275.26 1198.64L1319.57 1098.11C1321.24 1094.32 1323.63 1090.88 1326.59 1087.98L1367.03 1048.43L1367.13 1048.34L1367.22 1048.23L1450.17 956.248C1455.85 949.945 1459.46 942.051 1460.52 933.63L1467.32 879.499C1469.6 861.418 1488.39 850.4 1505.27 857.245Z" stroke="black" stroke-width="9"/>
  </g>
  <path d="M1478.84 1121.22L1248.42 1059.33L1131.19 1366.29L1386.93 1405.31L1478.84 1121.22Z" fill="#252525"/>
  <path d="M1251.79 1065.93L1138.76 1361.88L1383.14 1399.16L1471.82 1125.03L1251.79 1065.93Z" stroke="black" stroke-opacity="0.5" stroke-width="11"/>
  <path d="M587.5 285C811.975 285 991 432.648 991 611.5C991 701.811 945.719 833.016 871.856 941.783C797.845 1050.77 697.059 1134.5 587.5 1134.5C477.941 1134.5 377.155 1050.77 303.144 941.783C229.282 833.016 184 701.811 184 611.5C184 432.648 363.025 285 587.5 285Z" fill="url(#paint7_radial_17_18)" stroke="black" stroke-width="16"/>
  <path d="M582.5 287.5C742.677 287.5 887.306 305.68 991.561 334.854C1043.76 349.461 1085.24 366.661 1113.43 385.31C1142.03 404.23 1154.5 422.926 1154.5 440C1154.5 457.074 1142.03 475.77 1113.43 494.69C1085.24 513.339 1043.76 530.539 991.561 545.146C887.306 574.32 742.677 592.5 582.5 592.5C422.323 592.5 277.694 574.32 173.439 545.146C121.239 530.539 79.7574 513.339 51.5693 494.69C22.9701 475.77 10.5 457.074 10.5 440C10.5 422.926 22.9701 404.23 51.5693 385.31C79.7574 366.661 121.239 349.461 173.439 334.854C277.694 305.68 422.323 287.5 582.5 287.5Z" fill="url(#paint8_radial_17_18)" stroke="black" stroke-width="21"/>
  <path d="M330.772 20.0459C353.289 11.5475 376.106 9.55293 401.04 10.8711C426.252 12.204 452.812 16.8429 483.432 21.3242C513.786 25.7667 547.43 29.9248 585.505 29.9062C623.369 29.8877 657.57 25.7429 688.838 21.3359C720.352 16.8943 748.302 12.2913 774.867 10.9932C801.207 9.7061 825.425 11.737 848.956 20.2686C872.434 28.7807 895.903 44.0101 920.339 70.0859C937.026 87.8933 949.578 118.464 958.594 156.08C967.522 193.332 972.64 235.961 975.463 276.449C978.283 316.885 978.796 354.9 978.602 382.825C978.505 396.78 978.231 408.195 977.982 416.11C977.93 417.781 977.877 419.296 977.829 420.646C976.935 420.869 975.966 421.112 974.923 421.37C968.136 423.052 958.254 425.459 945.903 428.353C921.199 434.141 886.637 441.875 847.224 449.648C768.214 465.232 670.388 480.838 593.41 481.5C512.054 482.2 408.525 466.598 324.881 450.67C283.156 442.724 246.558 434.734 220.396 428.73C207.316 425.729 196.851 423.225 189.662 421.474C188.557 421.204 187.53 420.951 186.582 420.719C186.6 419.515 186.621 418.188 186.648 416.742C186.797 408.89 187.096 397.561 187.693 383.702C188.889 355.97 191.279 318.182 196.052 277.892C200.83 237.555 207.968 194.97 218.586 157.554C229.292 119.828 243.174 88.7094 260.654 70.0938C285.337 43.8088 308.267 28.5403 330.772 20.0459Z" fill="url(#paint9_radial_17_18)" stroke="black" stroke-width="21"/>
  </g>
  <defs>
  <filter id="filter0_n_17_18" x="0" y="0" width="1647.86" height="1405.31" filterUnits="userSpaceOnUse" color-interpolation-filters="sRGB">
  <feFlood flood-opacity="0" result="BackgroundImageFix"/>
  <feBlend mode="normal" in="SourceGraphic" in2="BackgroundImageFix" result="shape"/>
  <feTurbulence type="fractalNoise" baseFrequency="2 2" stitchTiles="stitch" numOctaves="3" result="noise" seed="4652" />
  <feColorMatrix in="noise" type="luminanceToAlpha" result="alphaNoise" />
  <feComponentTransfer in="alphaNoise" result="coloredNoise1">
  <feFuncA type="discrete" tableValues="1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 "/>
  </feComponentTransfer>
  <feComposite operator="in" in2="shape" in="coloredNoise1" result="noise1Clipped" />
  <feFlood flood-color="rgba(0, 0, 0, 0.25)" result="color1Flood" />
  <feComposite operator="in" in2="noise1Clipped" in="color1Flood" result="color1" />
  <feMerge result="effect1_noise_17_18">
  <feMergeNode in="shape" />
  <feMergeNode in="color1" />
  </feMerge>
  </filter>
  <clipPath id="paint2_diamond_17_18_clip_path"><path d="M1305.36 645.126L1331.55 724.938L1243.34 753.89L1217.14 674.078L1305.36 645.126Z" stroke-width="10"/></clipPath><filter id="filter1_f_17_18" x="1173.23" y="630.113" width="248.305" height="486.499" filterUnits="userSpaceOnUse" color-interpolation-filters="sRGB">
  <feFlood flood-opacity="0" result="BackgroundImageFix"/>
  <feBlend mode="normal" in="SourceGraphic" in2="BackgroundImageFix" result="shape"/>
  <feGaussianBlur stdDeviation="2" result="effect1_foregroundBlur_17_18"/>
  </filter>
  <filter id="filter2_f_17_18" x="1266.38" y="846.692" width="264.715" height="474.475" filterUnits="userSpaceOnUse" color-interpolation-filters="sRGB">
  <feFlood flood-opacity="0" result="BackgroundImageFix"/>
  <feBlend mode="normal" in="SourceGraphic" in2="BackgroundImageFix" result="shape"/>
  <feGaussianBlur stdDeviation="2" result="effect1_foregroundBlur_17_18"/>
  </filter>
  <radialGradient id="paint0_radial_17_18" cx="0" cy="0" r="1" gradientUnits="userSpaceOnUse" gradientTransform="translate(1312.97 795.848) rotate(105.401) scale(168 38.5)">
  <stop stop-color="#161616"/>
  <stop offset="1"/>
  </radialGradient>
  <radialGradient id="paint1_radial_17_18" cx="0" cy="0" r="1" gradientUnits="userSpaceOnUse" gradientTransform="translate(1320.98 684.223) rotate(161.83) scale(253.5 47)">
  <stop stop-color="#242424"/>
  <stop offset="1" stop-color="#2F2F2F"/>
  </radialGradient>
  <linearGradient id="paint2_diamond_17_18" x1="0" y1="0" x2="500" y2="500" gradientUnits="userSpaceOnUse">
  <stop stop-color="#684802"/>
  <stop offset="1" stop-color="#3C2901"/>
  </linearGradient>
  <linearGradient id="paint3_linear_17_18" x1="1567.83" y1="467.246" x2="1638.26" y2="459.889" gradientUnits="userSpaceOnUse">
  <stop stop-color="#5C5C5C"/>
  <stop offset="1" stop-color="#333333"/>
  </linearGradient>
  <linearGradient id="paint4_linear_17_18" x1="1514.09" y1="488.851" x2="1595.78" y2="549.919" gradientUnits="userSpaceOnUse">
  <stop stop-color="#5C5C5C"/>
  <stop offset="1" stop-color="#333333"/>
  </linearGradient>
  <radialGradient id="paint5_radial_17_18" cx="0" cy="0" r="1" gradientUnits="userSpaceOnUse" gradientTransform="translate(1228.64 768.222) rotate(74.6414) scale(154.735 51.0226)">
  <stop/>
  <stop offset="1" stop-color="#111010"/>
  </radialGradient>
  <radialGradient id="paint6_radial_17_18" cx="0" cy="0" r="1" gradientUnits="userSpaceOnUse" gradientTransform="translate(1472.85 983.119) rotate(111.727) scale(154.735 51.0226)">
  <stop/>
  <stop offset="1" stop-color="#111010"/>
  </radialGradient>
  <radialGradient id="paint7_radial_17_18" cx="0" cy="0" r="1" gradientUnits="userSpaceOnUse" gradientTransform="translate(587.5 709.75) rotate(90) scale(432.75 411.5)">
  <stop stop-color="#0B0B0B"/>
  <stop offset="1" stop-color="#0D0D0D"/>
  </radialGradient>
  <radialGradient id="paint8_radial_17_18" cx="0" cy="0" r="1" gradientUnits="userSpaceOnUse" gradientTransform="translate(582.5 440) rotate(90) scale(163 582.5)">
  <stop stop-color="#3D3D3D"/>
  <stop offset="1" stop-color="#1A1A1A"/>
  </radialGradient>
  <radialGradient id="paint9_radial_17_18" cx="0" cy="0" r="1" gradientUnits="userSpaceOnUse" gradientTransform="translate(582.569 246.012) rotate(90) scale(246.012 406.569)">
  <stop stop-color="#3D3D3D"/>
  <stop offset="1" stop-color="#1A1A1A"/>
  </radialGradient>
  </defs>
</svg>

<!-- speak bubble -->
<svg viewBox="0 0 3139 1256" fill="none" xmlns="http://www.w3.org/2000/svg">
  <path d="M696.115 216.13C627.545 314.584 697.994 449.5 817.973 449.5H3015C3069.95 449.5 3114.5 494.048 3114.5 549V1132C3114.5 1186.95 3069.95 1231.5 3015 1231.5H124C69.0477 1231.5 24.5 1186.95 24.5 1132V488.2C24.5 458.054 40.3017 430.14 66.0889 414.614L803.383 62.1123L696.115 216.13Z" fill="#242424" stroke="black" stroke-width="49"/>
</svg>

<div>
  It's like this, my dear sir, You're wasting your life
  cutting hair, lathering faces and swapping idle chitchat. When you're dead,
  it'll be as if you'd never existed. If you only had the time to lead the right
  kind of life, you'd be quite a different person. Time is all you need, right ?
</div>

<style>
  @import url('https://fonts.googleapis.com/css2?family=Cabin+Sketch:wght@400;700&display=swap');

  .chalk-font {
    font-family: "Cabin Sketch", sans-serif;
    font-weight: 400;
    font-style: normal; 
  }

  .mirror-glass {
    box-shadow: rgba(255, 255, 255, 0.04) 0px 6px 24px 0px, rgba(255, 255, 255, 0.08) 0px 0px 0px 1px;
    background-color: rgba(255, 255, 255, 0.005);
    backdrop-filter: blur(2px);
  }
</style>