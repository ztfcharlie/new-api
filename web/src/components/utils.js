import { API, showError } from '../helpers';

export async function getOAuthState() {
  let path = '/api/oauth/state';
  let affCode = localStorage.getItem('aff');
  if (affCode && affCode.length > 0) {
    path += `?aff=${affCode}`;
  }
  const res = await API.get(path);
  const { success, message, data } = res.data;
  if (success) {
    return data;
  } else {
    showError(message);
    return '';
  }
}
function geBurnCloudOAuthRedirectUri() {
  const protocol = window.location.protocol;
  const hostname = window.location.hostname;
  const port = window.location.port;
  const fullDomain = port ? `${hostname}:${port}` : hostname;
  return `${protocol}//${fullDomain}/oauth/burncloud`;
}

export async function onBurnCloudOAuthClicked(burncloud_client_id) {
  const state = await getOAuthState();
  if (!state) return;
  const redirect_uri = encodeURIComponent(geBurnCloudOAuthRedirectUri())
  window.open(
    `${import.meta.env.VITE_CONSOLE_DOMAIN}clogin?client_id=${burncloud_client_id}&state=${state}&redirect_uri=${redirect_uri}`,
  );
}
export async function onOIDCClicked(auth_url, client_id, openInNewTab = false) {
  const state = await getOAuthState();
  if (!state) return;
  const redirect_uri = `${window.location.origin}/oauth/oidc`;
  const response_type = "code";
  const scope = "openid profile email";
  const url = `${auth_url}?client_id=${client_id}&redirect_uri=${redirect_uri}&response_type=${response_type}&scope=${scope}&state=${state}`;
  if (openInNewTab) {
    window.open(url);
  } else
  {
    window.location.href = url;
  }
}

export async function onGitHubOAuthClicked(github_client_id) {
  const state = await getOAuthState();
  if (!state) return;
  window.open(
    `https://github.com/login/oauth/authorize?client_id=${github_client_id}&state=${state}&scope=user:email`,
  );
}

export async function onLinuxDOOAuthClicked(linuxdo_client_id) {
  const state = await getOAuthState();
  if (!state) return;
  window.open(
    `https://connect.linux.do/oauth2/authorize?response_type=code&client_id=${linuxdo_client_id}&state=${state}`,
  );
}

let channelModels = undefined;
export async function loadChannelModels() {
  const res = await API.get('/api/models');
  const { success, data } = res.data;
  if (!success) {
    return;
  }
  channelModels = data;
  localStorage.setItem('channel_models', JSON.stringify(data));
}

export function getChannelModels(type) {
  if (channelModels !== undefined && type in channelModels) {
    if (!channelModels[type]) {
      return [];
    }
    return channelModels[type];
  }
  let models = localStorage.getItem('channel_models');
  if (!models) {
    return [];
  }
  channelModels = JSON.parse(models);
  if (type in channelModels) {
    return channelModels[type];
  }
  return [];
}
