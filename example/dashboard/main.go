package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var collection *mongo.Collection

func main() {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	listenAddr := os.Getenv("LISTEN")
	if listenAddr == "" {
		listenAddr = ":8080"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	collection = client.Database("machineroom").Collection("nodes")

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/nodes", handleNodes)

	log.Printf("Dashboard listening on %s", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

func handleNodes(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, indexHTML)
}

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Machine Room — Node Dashboard</title>
<style>
  :root {
    --bg: #0f1117;
    --surface: #181b23;
    --surface2: #1e2130;
    --border: #2a2d3a;
    --text: #e1e4ed;
    --text-dim: #8b8fa3;
    --accent: #6c8cff;
    --accent-dim: #3d5afe;
    --green: #34d399;
    --yellow: #fbbf24;
    --red: #f87171;
    --orange: #fb923c;
    --radius: 8px;
  }
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Inter', sans-serif;
    background: var(--bg);
    color: var(--text);
    line-height: 1.5;
    min-height: 100vh;
  }
  .header {
    background: var(--surface);
    border-bottom: 1px solid var(--border);
    padding: 16px 32px;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .header h1 {
    font-size: 18px;
    font-weight: 600;
    letter-spacing: -0.02em;
  }
  .header h1 span { color: var(--accent); }
  .header-stats {
    display: flex;
    gap: 24px;
    font-size: 13px;
    color: var(--text-dim);
  }
  .header-stats .val { color: var(--text); font-weight: 600; }
  .container { max-width: 1400px; margin: 0 auto; padding: 24px 32px; }
  .search-bar {
    margin-bottom: 20px;
    position: relative;
  }
  .search-bar input {
    width: 100%;
    padding: 10px 16px 10px 40px;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    color: var(--text);
    font-size: 14px;
    outline: none;
    transition: border-color 0.15s;
  }
  .search-bar input:focus { border-color: var(--accent); }
  .search-bar svg {
    position: absolute;
    left: 14px;
    top: 50%;
    transform: translateY(-50%);
    color: var(--text-dim);
    width: 16px;
    height: 16px;
  }
  table {
    width: 100%;
    border-collapse: collapse;
    background: var(--surface);
    border-radius: var(--radius);
    overflow: hidden;
    border: 1px solid var(--border);
  }
  thead th {
    text-align: left;
    padding: 10px 16px;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--text-dim);
    background: var(--surface2);
    border-bottom: 1px solid var(--border);
    white-space: nowrap;
    cursor: pointer;
    user-select: none;
  }
  thead th:hover { color: var(--text); }
  thead th .sort-arrow { margin-left: 4px; font-size: 10px; }
  tbody tr {
    cursor: pointer;
    transition: background 0.1s;
    border-bottom: 1px solid var(--border);
  }
  tbody tr:last-child { border-bottom: none; }
  tbody tr:hover { background: var(--surface2); }
  tbody td {
    padding: 10px 16px;
    font-size: 13px;
    white-space: nowrap;
  }
  .status-dot {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    margin-right: 8px;
  }
  .status-dot.connected { background: var(--green); box-shadow: 0 0 6px var(--green); }
  .status-dot.provisioning { background: var(--yellow); box-shadow: 0 0 6px var(--yellow); }
  .tag {
    display: inline-block;
    padding: 2px 8px;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 500;
    background: rgba(108,140,255,0.12);
    color: var(--accent);
    margin-right: 4px;
  }
  .tag.os { background: rgba(52,211,153,0.12); color: var(--green); }
  .tag.role { background: rgba(251,191,36,0.12); color: var(--yellow); }
  .mem-bar {
    width: 80px;
    height: 6px;
    background: var(--border);
    border-radius: 3px;
    overflow: hidden;
    display: inline-block;
    vertical-align: middle;
    margin-right: 8px;
  }
  .mem-bar-fill {
    height: 100%;
    border-radius: 3px;
    transition: width 0.3s;
  }

  /* Expanded detail panel */
  .detail-row td {
    padding: 0 !important;
    background: var(--bg);
    cursor: default;
  }
  .detail-row:hover { background: var(--bg) !important; }
  .detail-panel {
    padding: 20px 24px;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
    gap: 16px;
  }
  .detail-card {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 16px;
  }
  .detail-card h3 {
    font-size: 12px;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--text-dim);
    margin-bottom: 12px;
    padding-bottom: 8px;
    border-bottom: 1px solid var(--border);
  }
  .detail-card dl {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 4px 16px;
    font-size: 13px;
  }
  .detail-card dt { color: var(--text-dim); white-space: nowrap; }
  .detail-card dd { color: var(--text); word-break: break-all; }
  .detail-card ul {
    list-style: none;
    font-size: 13px;
  }
  .detail-card li {
    padding: 6px 0;
    border-bottom: 1px solid var(--border);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .detail-card li:last-child { border-bottom: none; }
  .detail-card .agent-name { font-weight: 500; }
  .detail-card .agent-ver { color: var(--text-dim); font-size: 12px; }
  .machine-state {
    display: inline-block;
    padding: 2px 8px;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 600;
    background: rgba(52,211,153,0.12);
    color: var(--green);
  }
  .loading {
    text-align: center;
    padding: 60px;
    color: var(--text-dim);
    font-size: 14px;
  }
  .error {
    text-align: center;
    padding: 40px;
    color: var(--red);
    font-size: 14px;
  }
</style>
</head>
<body>
<div class="header">
  <h1><span>Machine Room</span> — Node Dashboard</h1>
  <div class="header-stats">
    <div>Nodes: <span class="val" id="node-count">—</span></div>
    <div>Last refresh: <span class="val" id="last-refresh">—</span></div>
  </div>
</div>
<div class="container">
  <div class="search-bar">
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
    <input type="text" id="search" placeholder="Filter nodes by hostname, IP, platform, role...">
  </div>
  <div id="content"><div class="loading">Loading nodes...</div></div>
</div>
<script>
let nodes = [];
let expanded = new Set();
let sortCol = 'hostname';
let sortAsc = true;

async function fetchNodes() {
  try {
    const resp = await fetch('/api/nodes');
    if (!resp.ok) throw new Error(resp.statusText);
    nodes = await resp.json();
    document.getElementById('node-count').textContent = nodes.length;
    document.getElementById('last-refresh').textContent = new Date().toLocaleTimeString();
    render();
  } catch (e) {
    document.getElementById('content').innerHTML = '<div class="error">Failed to load nodes: ' + e.message + '</div>';
  }
}

function get(obj, path) {
  return path.split('.').reduce((o, k) => o && o[k], obj);
}

function formatBytes(b) {
  if (!b) return '—';
  const gb = b / (1024*1024*1024);
  if (gb >= 1) return gb.toFixed(1) + ' GB';
  return (b / (1024*1024)).toFixed(0) + ' MB';
}

function formatUptime(secs) {
  if (!secs && secs !== 0) return '—';
  const d = Math.floor(secs / 86400);
  const h = Math.floor((secs % 86400) / 3600);
  const m = Math.floor((secs % 3600) / 60);
  if (d > 0) return d + 'd ' + h + 'h';
  if (h > 0) return h + 'h ' + m + 'm';
  return m + 'm';
}

function formatLastSeen(ts) {
  if (!ts) return {text: '—', stale: false};
  const d = new Date(ts);
  const now = Date.now();
  const ago = Math.floor((now - d.getTime()) / 1000);
  const stale = ago > 900;
  let text;
  if (ago < 60) text = ago + 's ago';
  else if (ago < 3600) text = Math.floor(ago / 60) + 'm ago';
  else if (ago < 86400) text = Math.floor(ago / 3600) + 'h ' + Math.floor((ago % 3600) / 60) + 'm ago';
  else text = Math.floor(ago / 86400) + 'd ago';
  return {text, stale};
}

function memColor(pct) {
  if (pct > 80) return 'var(--red)';
  if (pct > 60) return 'var(--orange)';
  if (pct > 40) return 'var(--yellow)';
  return 'var(--green)';
}

function sortNodes(list) {
  const extract = {
    'customer': n => {
      const ext = get(n, 'node.facts.machine_room.provisioning.extended_claims') || {};
      return (ext.customer || '').toLowerCase();
    },
    'hostname': n => (n.hostname || '').toLowerCase(),
    'ip': n => get(n, 'node.facts.network.default_ipv4') || '',
    'platform': n => {
      const hi = get(n, 'node.facts.host.info') || {};
      return (hi.platform || '') + ' ' + (hi.platformVersion || '');
    },
    'role': n => {
      const ext = get(n, 'node.facts.machine_room.provisioning.extended_claims') || {};
      return ext.role || '';
    },
    'memory': n => get(n, 'node.facts.memory.virtual.usedPercent') || 0,
    'uptime': n => get(n, 'node.facts.host.info.uptime') || 0,
    'agents': n => (get(n, 'node.agents') || []).length,
    'lastseen': n => n.timestamp ? new Date(n.timestamp).getTime() : 0,
    'status': n => get(n, 'node.status.provisioning_mode') ? 0 : 1,
  };
  const fn = extract[sortCol] || extract['hostname'];
  return [...list].sort((a, b) => {
    let va = fn(a), vb = fn(b);
    if (typeof va === 'string') { va = va.toLowerCase(); vb = vb.toLowerCase(); }
    if (va < vb) return sortAsc ? -1 : 1;
    if (va > vb) return sortAsc ? 1 : -1;
    return 0;
  });
}

function filterNodes(list) {
  const q = (document.getElementById('search').value || '').toLowerCase();
  if (!q) return list;
  return list.filter(n => {
    const hostname = (n.hostname || '').toLowerCase();
    const ip = (get(n, 'node.facts.network.default_ipv4') || '').toLowerCase();
    const hi = get(n, 'node.facts.host.info') || {};
    const plat = ((hi.platform || '') + ' ' + (hi.platformVersion || '')).toLowerCase();
    const ext = get(n, 'node.facts.machine_room.provisioning.extended_claims') || {};
    const role = (ext.role || '').toLowerCase();
    const customer = (ext.customer || '').toLowerCase();
    return hostname.includes(q) || ip.includes(q) || plat.includes(q) || role.includes(q) || customer.includes(q);
  });
}

function renderDetail(n) {
  const hi = get(n, 'node.facts.host.info') || {};
  const mem = get(n, 'node.facts.memory.virtual') || {};
  const swap = get(n, 'node.facts.memory.swap.info') || {};
  const net = get(n, 'node.facts.network') || {};
  const agents = get(n, 'node.agents') || [];
  const machines = get(n, 'node.machines') || [];
  const status = get(n, 'node.status') || {};
  const build = get(n, 'node.build_info') || {};
  const ext = get(n, 'node.facts.machine_room.provisioning.extended_claims') || {};
  const opts = get(n, 'node.facts.machine_room.options') || {};
  const addrs = (net.addresses || []).filter(a => a.address);
  const ifaces = (net.interfaces || []).filter(i => (i.flags || []).includes('up'));

  return '<tr class="detail-row"><td colspan="10"><div class="detail-panel">' +
    '<div class="detail-card"><h3>Host Information</h3><dl>' +
      '<dt>Hostname</dt><dd>' + (hi.hostname || '—') + '</dd>' +
      '<dt>OS</dt><dd>' + (hi.os || '—') + '</dd>' +
      '<dt>Platform</dt><dd>' + (hi.platform || '—') + ' ' + (hi.platformVersion || '') + '</dd>' +
      '<dt>Platform Family</dt><dd>' + (hi.platformFamily || '—') + '</dd>' +
      '<dt>Kernel</dt><dd>' + (hi.kernelVersion || '—') + '</dd>' +
      '<dt>Architecture</dt><dd>' + (hi.kernelArch || '—') + '</dd>' +
      '<dt>Virtualization</dt><dd>' + (hi.virtualizationSystem || '—') + ' (' + (hi.virtualizationRole || '—') + ')</dd>' +
      '<dt>Uptime</dt><dd>' + formatUptime(hi.uptime) + '</dd>' +
      '<dt>Processes</dt><dd>' + (hi.procs || '—') + '</dd>' +
      '<dt>Build Version</dt><dd>' + (build.version || '—') + '</dd>' +
    '</dl></div>' +

    '<div class="detail-card"><h3>Memory</h3><dl>' +
      '<dt>Total</dt><dd>' + formatBytes(mem.total) + '</dd>' +
      '<dt>Used</dt><dd>' + formatBytes(mem.used) + ' (' + (mem.usedPercent || 0).toFixed(1) + '%)</dd>' +
      '<dt>Available</dt><dd>' + formatBytes(mem.available) + '</dd>' +
      '<dt>Cached</dt><dd>' + formatBytes(mem.cached) + '</dd>' +
      '<dt>Buffers</dt><dd>' + formatBytes(mem.buffers) + '</dd>' +
      '<dt>Swap Total</dt><dd>' + formatBytes(swap.total) + '</dd>' +
      '<dt>Swap Used</dt><dd>' + formatBytes(swap.used) + ' (' + (swap.usedPercent || 0).toFixed(1) + '%)</dd>' +
    '</dl></div>' +

    '<div class="detail-card"><h3>Network</h3><dl>' +
      '<dt>IPv4</dt><dd>' + (net.default_ipv4 || '—') + '</dd>' +
      '<dt>IPv6</dt><dd>' + (net.default_ipv6 || '—') + '</dd>' +
    '</dl>' +
    (ifaces.length ? '<ul style="margin-top:12px">' + ifaces.map(i =>
      '<li><span class="agent-name">' + i.name + '</span><span class="agent-ver">MTU ' + i.mtu +
      (i.hardwareAddr ? ' • ' + i.hardwareAddr : '') +
      (i.addrs ? ' • ' + i.addrs.map(a => a.addr).join(', ') : '') +
      '</span></li>'
    ).join('') + '</ul>' : '') +
    '</div>' +

    '<div class="detail-card"><h3>Connection Status</h3><dl>' +
      '<dt>Identity</dt><dd>' + (status.identity || '—') + '</dd>' +
      '<dt>Connected To</dt><dd>' + (status.connected_server || '—') + '</dd>' +
      '<dt>Provisioning</dt><dd>' + (status.provisioning_mode ? 'Yes' : 'No') + '</dd>' +
      '<dt>Token Expires</dt><dd>' + (status.token_expires || '—') + '</dd>' +
      '<dt>Uptime</dt><dd>' + formatUptime(status.uptime) + '</dd>' +
      '<dt>Collectives</dt><dd>' + ((get(n, 'node.collectives') || []).join(', ') || '—') + '</dd>' +
    '</dl></div>' +

    '<div class="detail-card"><h3>Agents (' + agents.length + ')</h3><ul>' +
    agents.map(a =>
      '<li><span class="agent-name">' + a.name + '</span><span class="agent-ver">v' + a.version + '</span></li>'
    ).join('') +
    '</ul></div>' +

    '<div class="detail-card"><h3>Autonomous Agents (' + machines.length + ')</h3><ul>' +
    machines.map(m =>
      '<li><span class="agent-name">' + m.name + ' <span class="agent-ver">v' + m.version + '</span></span><span class="machine-state">' + m.state + '</span></li>'
    ).join('') +
    '</ul>' +
    (Object.keys(ext).length ? '<h3 style="margin-top:16px">Extended Claims</h3><dl>' +
      Object.entries(ext).map(([k,v]) => '<dt>' + k + '</dt><dd>' + v + '</dd>').join('') +
    '</dl>' : '') +
    '</div>' +

  '</div></td></tr>';
}

function render() {
  const filtered = filterNodes(nodes);
  const sorted = sortNodes(filtered);

  const cols = [
    ['status', 'Status'],
    ['customer', 'Customer'],
    ['hostname', 'Hostname'],
    ['ip', 'IP Address'],
    ['platform', 'Platform'],
    ['role', 'Role'],
    ['memory', 'Memory'],
    ['uptime', 'Uptime'],
    ['agents', 'Agents'],
    ['lastseen', 'Last Seen'],
  ];

  let html = '<table><thead><tr>';
  cols.forEach(([key, label]) => {
    const arrow = sortCol === key ? (sortAsc ? ' ▲' : ' ▼') : '';
    html += '<th data-col="' + key + '">' + label + '<span class="sort-arrow">' + arrow + '</span></th>';
  });
  html += '</tr></thead><tbody>';

  sorted.forEach(n => {
    const id = n._id;
    const hi = get(n, 'node.facts.host.info') || {};
    const ip = get(n, 'node.facts.network.default_ipv4') || '—';
    const prov = get(n, 'node.status.provisioning_mode');
    const memPct = get(n, 'node.facts.memory.virtual.usedPercent') || 0;
    const memTotal = get(n, 'node.facts.memory.virtual.total');
    const uptime = hi.uptime;
    const agentCount = (get(n, 'node.agents') || []).length;
    const ext = get(n, 'node.facts.machine_room.provisioning.extended_claims') || {};

    html += '<tr data-id="' + id + '">';
    html += '<td><span class="status-dot ' + (prov ? 'provisioning' : 'connected') + '"></span>' + (prov ? 'Provisioning' : 'Connected') + '</td>';
    html += '<td><span class="tag">' + (ext.customer || '—') + '</span></td>';
    html += '<td><strong>' + (n.hostname || '—') + '</strong></td>';
    html += '<td>' + ip + '</td>';
    html += '<td><span class="tag os">' + (hi.platform || '—') + ' ' + (hi.platformVersion || '') + '</span></td>';
    html += '<td>' + (ext.role ? '<span class="tag role">' + ext.role + '</span>' : '—') + '</td>';
    html += '<td><div class="mem-bar"><div class="mem-bar-fill" style="width:' + memPct.toFixed(0) + '%;background:' + memColor(memPct) + '"></div></div>' + memPct.toFixed(1) + '% of ' + formatBytes(memTotal) + '</td>';
    html += '<td>' + formatUptime(uptime) + '</td>';
    html += '<td>' + agentCount + '</td>';
    const ls = formatLastSeen(n.timestamp);
    html += '<td style="color:' + (ls.stale ? 'var(--red)' : 'var(--text)') + ';font-weight:' + (ls.stale ? '600' : 'normal') + '">' + ls.text + '</td>';
    html += '</tr>';

    if (expanded.has(id)) {
      html += renderDetail(n);
    }
  });

  html += '</tbody></table>';
  document.getElementById('content').innerHTML = html;

  // Bind events
  document.querySelectorAll('thead th').forEach(th => {
    th.addEventListener('click', () => {
      const col = th.dataset.col;
      if (sortCol === col) { sortAsc = !sortAsc; }
      else { sortCol = col; sortAsc = true; }
      render();
    });
  });

  document.querySelectorAll('tbody tr[data-id]').forEach(tr => {
    tr.addEventListener('click', () => {
      const id = tr.dataset.id;
      if (expanded.has(id)) { expanded.delete(id); }
      else { expanded.add(id); }
      render();
    });
  });
}

document.getElementById('search').addEventListener('input', render);
fetchNodes();
setInterval(fetchNodes, 30000);
</script>
</body>
</html>
`
