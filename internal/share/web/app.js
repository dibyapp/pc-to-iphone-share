const $ = id => document.getElementById(id);
let current = '', entries = [], requestId = 0, nextOffset = 0, total = 0;
const size = bytes => { if (bytes === 0) return '0 B'; const unit = Math.min(3, Math.floor(Math.log(bytes) / Math.log(1024))); return `${(bytes / 1024 ** unit).toFixed(unit ? 1 : 0)} ${['B', 'KB', 'MB', 'GB'][unit]}`; };
function showLogin() { $('login').hidden = false; $('browser').hidden = true; $('password').value = ''; $('files').replaceChildren(); entries = []; }
async function api(url, data) {
  const res = await fetch(url, data === undefined ? {} : { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) });
  const body = await res.json();
  if (!res.ok) { if (res.status === 401) showLogin(); throw new Error(body.error || 'Something went wrong.'); }
  return body;
}
function render() {
  const query = $('search').value.trim().toLocaleLowerCase();
  const shown = entries.filter(entry => entry.name.toLocaleLowerCase().includes(query));
  $('item-count').textContent = `${shown.length} shown · ${total} items`;
  $('files').replaceChildren();
  if (!shown.length) {
    const empty = document.createElement('div'); empty.className = 'empty';
    const title = document.createElement('strong'); title.textContent = query ? 'No matching files' : 'Your folder is ready';
    const description = document.createElement('p'); description.textContent = query ? 'Try a different search.' : 'Add files to the selected folder on your computer, then tap refresh.';
    empty.append(title, description); $('files').append(empty); return;
  }
  for (const entry of shown) {
    const relative = current ? `${current}/${entry.name}` : entry.name;
    const row = document.createElement(entry.directory ? 'button' : 'a'); row.className = 'file-row';
    if (entry.directory) { row.type = 'button'; row.addEventListener('click', () => load(relative)); }
    else { row.href = '/download?path=' + encodeURIComponent(relative); row.download = entry.name; }
    const icon = document.createElement('span'); icon.className = 'file-icon ' + (entry.directory ? 'folder' : ''); icon.textContent = entry.directory ? '▰' : '↓';
    const text = document.createElement('span'); text.className = 'file-text';
    const name = document.createElement('strong'); name.textContent = entry.name;
    const detail = document.createElement('small'); detail.textContent = entry.directory ? 'Folder' : new Date(entry.modified).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
    text.append(name, detail);
    const meta = document.createElement('span'); meta.className = 'file-meta'; meta.textContent = entry.directory ? '→' : size(entry.size);
    row.append(icon, text, meta); $('files').append(row);
  }
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
  const id = ++requestId;
  $('browser-error').textContent = ''; $('refresh').disabled = true;
  try {
    const data = await api('/api/list?path=' + encodeURIComponent(folder) + '&offset=' + (more ? nextOffset : 0));
    if (id !== requestId) return;
    current = data.path; entries = more ? entries.concat(data.entries) : data.entries;
    nextOffset = data.nextOffset; total = data.total;
    $('load-more').hidden = !data.hasMore;
    $('login').hidden = true; $('browser').hidden = false; $('folder-name').textContent = data.name;
    if (!more) $('search').value = ''; crumbs(); render();
    if (data.truncated) $('browser-error').textContent = 'This folder exceeds 10,000 entries. Split large collections into subfolders to see all files.';
  } catch (error) { if (!$('browser').hidden) $('browser-error').textContent = error.message; else if (error.message !== 'Please sign in.') { showLogin(); $('login-error').textContent = error.message; } }
  finally { if (id === requestId) $('refresh').disabled = false; }
}
$('login-form').addEventListener('submit', async event => {
  event.preventDefault(); $('login-error').textContent = ''; $('sign-in').disabled = true; $('sign-in').textContent = 'Unlocking…';
  const password = $('password').value;
  try { await api('/api/login', { password }); $('password').value = ''; await load(''); }
  catch (error) { $('login-error').textContent = error.message; }
  finally { $('sign-in').disabled = false; $('sign-in').textContent = 'Unlock folder →'; }
});
$('show-password').addEventListener('click', () => { const reveal = $('password').type === 'password'; $('password').type = reveal ? 'text' : 'password'; $('show-password').textContent = reveal ? 'Hide' : 'Show'; $('show-password').setAttribute('aria-label', reveal ? 'Hide password' : 'Show password'); });
$('logout').addEventListener('click', async () => { try { await api('/api/logout', {}); requestId++; showLogin(); } catch (error) { $('browser-error').textContent = error.message; } });
$('search').addEventListener('input', render);
$('refresh').addEventListener('click', () => load());
$('load-more').addEventListener('click', () => load(current, true));
load();
