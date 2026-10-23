const $ = id => document.getElementById(id);
let current = '', entries = [], requestId = 0, nextOffset = 0, total = 0, selected = new Set(), searchTimer, detailId = 0;
let listController;
const size = bytes => { if (!bytes) return '0 B'; const unit = Math.min(4, Math.floor(Math.log(bytes) / Math.log(1024))); return `${(bytes / 1024 ** unit).toFixed(unit ? 1 : 0)} ${['B', 'KB', 'MB', 'GB', 'TB'][unit]}`; };
function clearDetails() { detailId++; $('details-panel').hidden = true; $('detail-text').textContent = ''; }
function showLogin() { requestId++; listController?.abort(); clearDetails(); $('login').hidden = false; $('browser').hidden = true; $('password').value = ''; $('files').replaceChildren(); entries = []; selected.clear(); }
async function api(url, data, signal) {
  const res = await fetch(url, data === undefined ? { signal } : { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data), signal });
  const body = await res.json();
  if (!res.ok) { if (res.status === 401) showLogin(); throw new Error(body.error || 'Something went wrong.'); }
  return body;
}
function updateSelection() {
  $('selection-count').textContent = `${selected.size} selected`;
  $('zip-download').hidden = !selected.size;
  const params = new URLSearchParams(); selected.forEach(name => params.append('path', name));
  $('zip-download').href = '/download.zip?' + params;
}
async function inspectFile(relative, mode) {
  const id = ++detailId; $('details-panel').hidden = false; $('detail-title').textContent = mode === 'preview' ? 'Text preview' : 'SHA-256 checksum';
  $('detail-text').textContent = 'Reading…';
  try {
    const data = await api(`/api/${mode}?path=` + encodeURIComponent(relative));
    if (id !== detailId || $('browser').hidden) return;
    $('detail-text').textContent = mode === 'preview' ? data.text + (data.truncated ? '\n\n[Preview limited to 256 KiB]' : '') : `${data.sha256}\n${size(data.size)} · Compare this hash with the downloaded file.`;
  } catch (error) { if (id === detailId && !$('browser').hidden) $('detail-text').textContent = error.message; }
}
function render() {
  $('item-count').textContent = `${entries.length} shown · ${total} matching items`;
  $('files').replaceChildren();
  if (!entries.length) {
    const empty = document.createElement('div'); empty.className = 'empty';
    const title = document.createElement('strong'); title.textContent = $('search').value || $('file-type').value !== 'all' ? 'No matching files' : 'Your folder is ready';
    const description = document.createElement('p'); description.textContent = 'Change your filters, or add files on the computer and refresh.';
    empty.append(title, description); $('files').append(empty);
  }
  for (const entry of entries) {
    const relative = current ? `${current}/${entry.name}` : entry.name;
    const item = document.createElement('div'); item.className = 'file-item';
    const row = document.createElement(entry.directory ? 'button' : 'a'); row.className = 'file-row';
    if (entry.directory) { row.type = 'button'; row.addEventListener('click', () => load(relative)); }
    else { row.href = '/download?path=' + encodeURIComponent(relative); row.download = entry.name; }
    const icon = document.createElement('span'); icon.className = 'file-icon ' + (entry.directory ? 'folder' : ''); icon.textContent = entry.directory ? '▰' : '↓';
    const text = document.createElement('span'); text.className = 'file-text';
    const name = document.createElement('strong'); name.textContent = entry.name;
    const detail = document.createElement('small'); detail.textContent = entry.directory ? 'Folder' : new Date(entry.modified).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
    text.append(name, detail);
    const meta = document.createElement('span'); meta.className = 'file-meta'; meta.textContent = entry.directory ? '→' : size(entry.size);
    row.append(icon, text, meta);
    if (!entry.directory) {
      const check = document.createElement('input'); check.type = 'checkbox'; check.className = 'file-check'; check.setAttribute('aria-label', 'Select ' + entry.name); check.checked = selected.has(relative);
      check.addEventListener('change', () => {
        if (check.checked && selected.size >= 64) { check.checked = false; $('browser-error').textContent = 'Select at most 64 files per ZIP.'; return; }
        if (check.checked) selected.add(relative); else selected.delete(relative); updateSelection();
      });
      item.append(check);
    }
    item.append(row);
    if (!entry.directory) {
      const actions = document.createElement('div'); actions.className = 'file-actions';
      for (const [mode, label] of [['preview', 'Preview'], ['checksum', 'SHA-256']]) {
        if (mode === 'preview' && !/\.(txt|md|csv|json|log|ya?ml|xml)$/i.test(entry.name)) continue;
        const button = document.createElement('button'); button.type = 'button'; button.className = 'quiet'; button.textContent = label;
        button.addEventListener('click', () => inspectFile(relative, mode)); actions.append(button);
      }
      item.append(actions);
    }
    $('files').append(item);
  }
  updateSelection();
}
function crumbs() {
  $('breadcrumbs').replaceChildren();
  const parts = current ? current.split('/') : [];
  ['', ...parts].forEach((part, index) => {
    const button = document.createElement('button'); button.textContent = index ? part : 'Shared folder'; button.type = 'button';
    button.addEventListener('click', () => load(parts.slice(0, index).join('/')));
    if (index) { const divider = document.createElement('span'); divider.textContent = '/'; $('breadcrumbs').append(divider); }
    $('breadcrumbs').append(button);
  });
}
async function load(folder = current, more = false) {
  clearTimeout(searchTimer);
  const id = ++requestId; listController?.abort(); listController = new AbortController();
  if (folder !== current) { $('search').value = ''; selected.clear(); }
  $('browser-error').textContent = ''; $('refresh').disabled = true; $('load-more').disabled = true;
  try {
    const params = new URLSearchParams({ path: folder, offset: more ? nextOffset : 0, limit: $('page-size').value, q: $('search').value, type: $('file-type').value, sort: $('sort').value, direction: $('direction').value });
    const data = await api('/api/list?' + params, undefined, listController.signal);
    if (id !== requestId) return;
    current = data.path; entries = more ? entries.concat(data.entries) : data.entries;
    if (!more) { selected.clear(); clearDetails(); }
    nextOffset = data.nextOffset; total = data.total; $('load-more').hidden = !data.hasMore;
    $('login').hidden = true; $('browser').hidden = false; $('folder-name').textContent = data.name;
    $('expiry').textContent = data.expiresAt && !data.expiresAt.startsWith('0001') ? 'Sharing ends at ' + new Date(data.expiresAt).toLocaleTimeString() : '';
    crumbs(); render();
    if (data.truncated) $('browser-error').textContent = 'Search covers the first 10,000 entries in this folder. Split larger collections into subfolders.';
  } catch (error) {
    if (id !== requestId || error.name === 'AbortError') return;
    if (!$('browser').hidden) $('browser-error').textContent = error.message;
    else if (error.message !== 'Please sign in.') { showLogin(); $('login-error').textContent = error.message; }
  } finally { if (id === requestId) { $('refresh').disabled = false; $('load-more').disabled = false; } }
}
$('login-form').addEventListener('submit', async event => {
  event.preventDefault(); $('login-error').textContent = ''; $('sign-in').disabled = true; $('sign-in').textContent = 'Unlocking…';
  const password = $('password').value;
  try { await api('/api/login', { password }); $('password').value = ''; await load(''); }
  catch (error) { $('login-error').textContent = error.message; }
  finally { $('sign-in').disabled = false; $('sign-in').textContent = 'Unlock folder →'; }
});
$('show-password').addEventListener('click', () => { const reveal = $('password').type === 'password'; $('password').type = reveal ? 'text' : 'password'; $('show-password').textContent = reveal ? 'Hide' : 'Show'; $('show-password').setAttribute('aria-label', reveal ? 'Hide password' : 'Show password'); });
$('logout').addEventListener('click', async () => { try { await api('/api/logout', {}); showLogin(); } catch (error) { $('browser-error').textContent = error.message; } });
$('search').addEventListener('input', () => { clearTimeout(searchTimer); searchTimer = setTimeout(() => load(), 300); });
for (const id of ['file-type', 'sort', 'direction', 'page-size']) $(id).addEventListener('change', () => load());
$('refresh').addEventListener('click', () => load());
$('load-more').addEventListener('click', () => load(current, true));
$('select-page').addEventListener('click', () => { entries.filter(e => !e.directory).forEach(e => { if (selected.size < 64) selected.add(current ? current + '/' + e.name : e.name); }); render(); });
$('clear-selection').addEventListener('click', () => { selected.clear(); render(); });
$('close-details').addEventListener('click', clearDetails);
$('speed-test').addEventListener('click', async () => {
  $('speed-test').disabled = true; $('speed-result').textContent = 'Measuring 16 MiB of synthetic data…';
  const controller = new AbortController(); const timer = setTimeout(() => controller.abort(), 30000);
  try {
    const start = performance.now(); const response = await fetch('/api/speed-test', { signal: controller.signal });
    if (!response.ok) { const body = await response.json(); if (response.status === 401) showLogin(); throw new Error(body.error); }
    const bytes = (await response.arrayBuffer()).byteLength;
    if (bytes !== 16 * 1024 * 1024) throw new Error('Speed test interrupted. Try again.');
    const seconds = Math.max(0.001, (performance.now() - start) / 1000);
    if (!$('browser').hidden) $('speed-result').textContent = `${(bytes / seconds / 1e6).toFixed(1)} MB/s · ${(bytes * 8 / seconds / 1e6).toFixed(0)} Mbps. Local test only; file speed can differ.`;
  } catch (error) { $('speed-result').textContent = error.name === 'AbortError' ? 'Speed test timed out.' : error.message; }
  finally { clearTimeout(timer); $('speed-test').disabled = false; }
});
load();
