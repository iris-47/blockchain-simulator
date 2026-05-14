async function fetchJSON(url, opt) {
  const res = await fetch(url, opt);
  return res.json();
}

async function refresh() {
  const status = await fetchJSON('/api/status');
  const metrics = await fetchJSON('/api/metrics');
  document.getElementById('status').textContent = JSON.stringify(status.data || status, null, 2);
  document.getElementById('metrics').textContent = JSON.stringify(metrics.data || metrics, null, 2);
}

document.getElementById('refreshBtn').onclick = refresh;
document.getElementById('sendTxBtn').onclick = () => fetchJSON('/api/tx/send', { method: 'POST' }).then(refresh);
document.getElementById('startLoadBtn').onclick = () => fetchJSON('/api/test/throughput', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ tps: 100000, durationSec: 180 })
}).then(refresh);
document.getElementById('forgeBtn').onclick = () => fetchJSON('/api/control/forge-spacetime', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ shardId: 0, nodeId: 0, timeOffsetSec: 600, fakeLocation: 'fake-geo' })
}).then(refresh);

refresh();
setInterval(refresh, 2000);
