const $ = id => document.getElementById(id);
const token = location.hash.slice(1);
history.replaceState(null, '', '/');
let state;
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
$('stop').addEventListener('click', async () => {
  if (!confirm('Stop sharing? Active downloads have up to 30 seconds to finish.')) return;
  try { await request('/api/stop', 'POST'); $('stop').disabled = true; $('status').textContent = 'Sharing is stopping. This dashboard will disconnect.'; }
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
  } catch (error) { $('error').textContent = error.message; }
})();
