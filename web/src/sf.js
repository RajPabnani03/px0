// web/src/sf.js
import { $, api } from './state.js';

/* Status-bar chip for the workspace's default Salesforce org, backed by the
   sf CLI (/api/sf/default). Hidden entirely when sf is not installed, so the
   chip costs nothing on machines that are not Salesforce shops. The dot is
   the org's connected status; a click re-runs the sf calls, so a re-auth or
   `sf config set target-org` shows up without a reload. */

const chip = () => $('#st-sf');

export function initSfOrg() {
  const el = chip();
  if (!el) return;
  el.addEventListener('click', () => refreshSfOrg());
  el.addEventListener('keydown', e => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      refreshSfOrg();
    }
  });
  refreshSfOrg();
}

export async function refreshSfOrg() {
  const el = chip();
  if (!el || el.dataset.state === 'starting') return;
  el.dataset.state = 'starting';
  try {
    draw(await api('/api/sf/default'));
  } catch (e) {
    draw({ installed: true, error: e.message });
  }
}

function draw(d) {
  const el = chip();
  if (!el) return;
  if (!d.installed) {
    el.hidden = true;
    return;
  }
  el.hidden = false;
  const org = d.org || {};
  const name = org.alias || org.username || d.value || '';
  el.dataset.state = d.error ? 'failed' : org.connectedStatus === 'Connected' ? 'ready' : 'off';
  el.textContent = d.error ? 'sf: error' : `sf: ${name || 'no org'}`;
  el.title = [org.username, org.instanceUrl, org.connectedStatus, d.error, 'Click to refresh']
    .filter(Boolean).join('\n');
}
