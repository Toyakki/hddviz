function Plot() {
  Transform(chartData, opts, 0);
  BuildPie(runningData, opts, 0);
}

function BuildPie(finaldata, options, level) {
  const width = 600;
  const height = 400;
  const radius = Math.min(400, height) / 2;

  // Create SVG container
  const chartContainer = document.getElementById('chart');
  const svg = createSVGElement('svg', {
    width: width,
    height: height,
    style: 'display: block; margin: 0 auto;'
  });

  const g = createSVGElement('g', {
    transform: `translate(${width / 2}, ${height / 2})`
  });
  svg.appendChild(g);
  chartContainer.appendChild(svg);

  // Calculate pie data
  const pieData = calculatePie(runningData);

  // Draw pie slices
  pieData.forEach((slice, index) => {
    const path = createSVGElement('path', {
      d: arcPath(slice, radius),
      fill: runningColors[index] || '#ccc',
      opacity: slice.data.op,
      style: 'cursor: pointer; transition: opacity 0.2s;',
      class: 'pie-slice'
    });

    // Animate path
    animatePath(path, slice, radius);

    // Tooltip events
    path.addEventListener('mouseover', () => {
      showTooltip(slice.data);
    });

    path.addEventListener('mouseout', () => {
      hideTooltip();
    });

    // Click to drill down
    path.addEventListener('click', () => {
      chartContainer.innerHTML = '';
      if (level == 0) {
        const orderedChartData = chartData.sort((a, b) => 
          parseFloat(b[opts.num]) - parseFloat(a[opts.num])
        );
        Transform(orderedChartData, opts, 1, slice.data.room);
        BuildPie(runningData, opts, 1);
      } else {
        Transform(chartData, opts, 0);
        BuildPie(runningData, opts, 0);
      }
    });

    g.appendChild(path);
  });

  // Draw legend
  drawLegend(g, pieData, width, runningData.length);
}

function calculatePie(data) {
  const total = data.reduce((sum, d) => sum + d.energy, 0);
  let currentAngle = -Math.PI / 2;
  const slices = [];

  data.forEach(d => {
    const sliceAngle = (d.energy / total) * 2 * Math.PI;
    slices.push({
      data: d,
      startAngle: currentAngle,
      endAngle: currentAngle + sliceAngle,
      value: d.energy
    });
    currentAngle += sliceAngle;
  });

  return slices;
}

function arcPath(slice, radius) {
  const outerRadius = radius - 50;

  const x1 = outerRadius * Math.cos(slice.startAngle);
  const y1 = outerRadius * Math.sin(slice.startAngle);
  const x2 = outerRadius * Math.cos(slice.endAngle);
  const y2 = outerRadius * Math.sin(slice.endAngle);

  const largeArc = slice.endAngle - slice.startAngle > Math.PI ? 1 : 0;

  return [
    `M ${x1} ${y1}`,
    `A ${outerRadius} ${outerRadius} 0 ${largeArc} 1 ${x2} ${y2}`,
    `L 0 0`,
    'Z'
  ].join(' ');
}

function animatePath(path, slice, radius) {
  let progress = 0;
  const duration = 1000;
  const startTime = Date.now();

  function animate() {
    progress = Math.min((Date.now() - startTime) / duration, 1);
    
    const currentSlice = {
      ...slice,
      startAngle: slice.startAngle * progress - (Math.PI / 2),
      endAngle: slice.endAngle * progress - (Math.PI / 2)
    };

    path.setAttribute('d', arcPath(currentSlice, radius));

    if (progress < 1) {
      requestAnimationFrame(animate);
    }
  }

  animate();
}

function drawLegend(g, pieData, width, dataLength) {
  const legendOffsetY = -(dataLength * 10);

  pieData.forEach((slice, i) => {
    const legend = createSVGElement('g', {
      class: 'legend',
      transform: `translate(60, ${legendOffsetY + i * 28})`
    });

    const rect = createSVGElement('rect', {
      x: width / 3,
      y: 0,
      width: 18,
      height: 18,
      fill: runningColors[i] || '#ccc',
      opacity: slice.data.op
    });

    const text = createSVGElement('text', {
      x: (width / 2) - 115,
      y: 9,
      'text-anchor': 'end',
      'font-size': '12px',
      fill: '#000'
    });
    text.textContent = slice.data.label;

    legend.appendChild(rect);
    legend.appendChild(text);
    g.appendChild(legend);
  });
}

function showTooltip(data) {
  const tooltip = document.getElementById('tooltip');
  tooltip.style.opacity = '1';
  tooltip.style.display = 'block';
  document.getElementById('section').textContent = data.label;
  document.getElementById('value').textContent = data.energy;
}

/**
 * Hide tooltip
 */
function hideTooltip() {
  const tooltip = document.getElementById('tooltip');
  tooltip.style.opacity = '0';
  tooltip.style.display = 'none';
}


function createSVGElement(tag, attributes = {}) {
  const element = document.createElementNS('http://www.w3.org/2000/svg', tag);
  Object.keys(attributes).forEach(key => {
    element.setAttribute(key, attributes[key]);
  });
  return element;
}

/**
 * Transform data based on level (drill down or not)
 */
function Transform(chartdata, options, level, filter) {
  let result = [];
  let resultColors = [];
  let opcounter = 0;
  let counter = 0;

  if (level == 0) {
    for (let i in chartdata) {
      let hasMatch = false;
      for (let index = 0; index < result.length; ++index) {
        if (result[index][opts.rm] == chartdata[i][opts.rm]) {
          result[index][opts.num] = result[index][opts.num] + chartdata[i][opts.num];
          hasMatch = true;
          break;
        }
      }

      if (!hasMatch) {
        let ditem = {};
        ditem[opts.rm] = chartdata[i][opts.rm];
        ditem[opts.num] = chartdata[i][opts.num];
        ditem["label"] = chartdata[i][opts.rm];
        ditem["op"] = 1;
        result.push(ditem);

        resultColors[counter] = opts.color != undefined ? opts.color[chartdata[i][opts.rm]] : "";
        counter += 1;
      }
    }
  } else {
    for (let i in chartdata) {
      if (chartdata[i].room == filter) {
        let newobj = {};
        newobj.room = chartdata[i].room;
        newobj.appliance = chartdata[i].appliance;
        newobj.energy = chartdata[i].energy;
        newobj["label"] = chartdata[i][opts.ap];
        newobj["op"] = 1.0 - parseFloat("0." + opcounter);
        result.push(newobj);

        resultColors[counter] = opts.color[chartdata[i][opts.rm]];
        opcounter += 2;
        counter += 1;
      }
    }
  }

  runningColors = resultColors;
  runningData = result;
}

let opts = {
  "captions": {
    "Living": "Living",
    "Bedroom": "Bedroom",
    "Bathroom": "Bathroom",
    "Kitchen": "Kitchen"
  },
  "color": {
    "Living": "#660033",
    "Bedroom": "#0066ff",
    "Bathroom": "#ffaa00",
    "Kitchen": "#009933"
  },
  "rm": "room",
  "ap": "appliance",
  "num": "energy"
}

let chartData = [
  { "room": "Living", "appliance": "Lighting", "energy": 300 },
  { "room": "Living", "appliance": "TV", "energy": 100 },
  { "room": "Living", "appliance": "Heating", "energy": 400 },
  { "room": "Bedroom", "appliance": "Lighting1", "energy": 190 },
  { "room": "Bedroom", "appliance": "Lighting2", "energy": 55 },
  { "room": "Bedroom", "appliance": "Heating", "energy": 240 },
  { "room": "Bathroom", "appliance": "Lighting", "energy": 80 },
  { "room": "Bathroom", "appliance": "Hot Water", "energy": 390 },
  { "room": "Bathroom", "appliance": "Fan", "energy": 50 },
  { "room": "Kitchen", "appliance": "Dishwasher", "energy": 120 },
  { "room": "Kitchen", "appliance": "Stove", "energy": 350 },
  { "room": "Kitchen", "appliance": "Fridge", "energy": 110 },
  { "room": "Kitchen", "appliance": "Lighting", "energy": 100 }
];

let runningData = [];
let runningColors = [];