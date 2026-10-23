const $ = id => document.getElementById(id);
const token = location.hash.slice(1);
history.replaceState(null, '', '/');
let state;
let paused = false, statsTimer;
async function refreshStats() {
  try {
    const stats = await (await request('/api/stats')).json(); paused = stats.paused;
    const expiry = stats.expiresAt && !stats.expiresAt.startsWith('0001') ? ' · Ends ' + new Date(stats.expiresAt).toLocaleTimeString() : '';
    $('transfer-stats').textContent = `${paused ? 'Paused' : 'Sharing'} · ${stats.active}/${stats.maxDownloads} active operations · ${stats.completed} successful download responses · ${(stats.bytes / 1048576).toFixed(1)} MiB sent · ${stats.sessions} signed-in sessions${expiry}`;
    $('pause').textContent = paused ? 'Resume sharing' : 'Pause new access';
  } catch (error) { $('transfer-stats').textContent = 'Dashboard unavailable: ' + error.message; clearInterval(statsTimer); }
}
async function request(route, method = 'GET') {
  const response = await fetch(route, { method, headers: { Authorization: 'Bearer ' + token } });
  if (!response.ok) { const body = await response.json(); throw new Error(body.error || 'Request failed.'); }
  return response;
}
async function download(route, name) {
  try {
    const response = await request(route);
    const url = URL.createObjectURL(await response.blob());
    const link = document.createElement('a'); link.href = url; link.download = name; link.click();
    setTimeout(() => URL.revokeObjectURL(url), 10000);
  } catch (error) { $('error').textContent = error.message; }
}
async function copy(value) {
  try { await navigator.clipboard.writeText(value); $('status').textContent = 'Copied to clipboard.'; }
  catch { $('status').textContent = 'Select and copy the value above.'; }
}
$('copy-url').addEventListener('click', () => copy(state.url));
$('copy-password').addEventListener('click', () => copy(state.password));
$('download-profile').addEventListener('click', () => download('/api/profile', 'PC-to-iPhone-Share.mobileconfig'));
$('download-cert').addEventListener('click', () => download('/api/certificate', 'PC-to-iPhone-Share.cer'));
$('pause').addEventListener('click', async event => {
  event.preventDefault();
  $('pause').disabled = true;
  try { await request(paused ? '/api/resume' : '/api/pause', 'POST'); await refreshStats(); }
  catch (error) { $('error').textContent = error.message; }
  finally { $('pause').disabled = false; }
});
$('revoke').addEventListener('click', async event => {
  event.preventDefault();
  if (!confirm('Sign out every device? Active downloads may finish. The password stays unchanged.')) return;
  try { await request('/api/revoke', 'POST'); $('status').textContent = 'Devices signed out. The password is unchanged.'; await refreshStats(); }
  catch (error) { $('error').textContent = error.message; }
});
$('stop').addEventListener('click', async event => {
  event.preventDefault();
  if (!confirm('Stop sharing? Active downloads have up to 30 seconds to finish.')) return;
  try { await request('/api/stop', 'POST'); clearInterval(statsTimer); $('stop').disabled = true; $('status').textContent = 'Sharing is stopping. This dashboard will disconnect.'; }
  catch (error) { $('error').textContent = error.message; }
});
(async () => {
  try {
    if (!token) throw new Error('Open the full dashboard URL printed in the app console. Its private access token is kept only in this tab.');
    state = await (await request('/api/state')).json();
    $('folder').textContent = state.folder; $('phone-url').textContent = state.url;
    $('password').textContent = state.password || 'Existing password unchanged. Restart with --reset-password if you lost it.';
    $('copy-password').hidden = !state.password; $('fingerprint').textContent = state.fingerprint;
    $('version').textContent = 'PC to iPhone Share ' + state.version; $('settings').hidden = false;
    await refreshStats(); statsTimer = setInterval(refreshStats, 5000);
  } catch (error) { $('error').textContent = error.message; }
})();
