const $ = (id) => document.getElementById(id);

async function call(url, method = 'GET', body) {
  const res = await fetch(url, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: body ? JSON.stringify(body) : undefined
  });
  return await res.json();
}

$('startNet').onclick = () => call('/api/network/start', 'POST');
$('stopNet').onclick = () => call('/api/network/stop', 'POST');
$('throughput').onclick = () => call('/api/tx/throughput', 'POST', {
  rate: Number($('rate').value),
  durationSeconds: Number($('duration').value)
});

$('forge').onclick = () => call(`/api/node/forge-spacetime?shard=${$('shard').value}&node=${$('node').value}`, 'POST');
$('byzOn').onclick = () => call(`/api/node/byzantine?shard=${$('shard').value}&node=${$('node').value}&enabled=true&behavior=broadcast_bad_block`, 'POST');
$('byzOff').onclick = () => call(`/api/node/byzantine?shard=${$('shard').value}&node=${$('node').value}&enabled=false`, 'POST');

$('queryBlocks').onclick = async () => {
  const data = await call(`/api/node/blocks?shard=${$('shard').value}&node=${$('node').value}&start=${$('start').value}&end=${$('end').value}`);
  $('blocks').textContent = JSON.stringify(data, null, 2);
};

async function refresh() {
  try {
    const status = await call('/api/status');
    $('status').textContent = JSON.stringify(status, null, 2);
    const metrics = await call('/api/metrics');
    $('metrics').textContent = JSON.stringify(metrics, null, 2);
  } catch (e) {}
}
setInterval(refresh, 1000);
refresh();
