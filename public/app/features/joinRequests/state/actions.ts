import { getBackendSrv } from '@grafana/runtime';
import { contextSrv } from 'app/core/core';
import { AccessControlAction, createAsyncThunk, JoinRequester } from 'app/types';

export const fetchJoinRequesters = createAsyncThunk('users/fetchJoinRequesters', async () => {
  if (!contextSrv.hasPermission(AccessControlAction.OrgUsersAdd)) {
    return [];
  }

  const joinRequesters: JoinRequester[] = await getBackendSrv().get('/api/org/joinRequests');
  return joinRequesters;
});

export const rejectJoinRequest = createAsyncThunk('users/rejectJoinRequest', async (requestId: Number) => {
  await getBackendSrv().patch(`/api/org/joinRequest/${requestId}/reject`, {});
  return requestId.toString();
});

export const approveJoinRequest = createAsyncThunk('users/approveJoinRequest', async (requestId: Number) => {
  await getBackendSrv().patch(`/api/org/joinRequest/${requestId}/approve`, {});
  return requestId.toString();
});

