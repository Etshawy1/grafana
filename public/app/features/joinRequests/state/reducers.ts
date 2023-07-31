import { createEntityAdapter, createSlice } from '@reduxjs/toolkit';

import { JoinRequester } from 'app/types';

import { fetchJoinRequesters, rejectJoinRequest, approveJoinRequest, fetchActiveOrg } from './actions';

export type Status = 'idle' | 'loading' | 'succeeded' | 'failed';

const joinRequestsAdapter = createEntityAdapter({ selectId: (joinRequester: JoinRequester) => joinRequester.id });

export const selectors = joinRequestsAdapter.getSelectors();
export const initialState = joinRequestsAdapter.getInitialState<{ status: Status }>({ status: 'idle' });


const joinRequestsSlice = createSlice({
  name: 'joinRequests',
  initialState,
  reducers: {},
  extraReducers: (builder) => {
    builder
      .addCase(fetchJoinRequesters.pending, (state) => {
        state.status = 'loading';
      })
      .addCase(fetchJoinRequesters.fulfilled, (state, { payload: joinRequests }) => {
        joinRequestsAdapter.setAll(state, joinRequests);
        state.status = 'succeeded';
      })
      .addCase(fetchJoinRequesters.rejected, (state) => {
        state.status = 'failed';
      })
      .addCase(rejectJoinRequest.fulfilled, (state, { payload: id }) => {
        joinRequestsAdapter.removeOne(state, id);
        state.status = 'succeeded';
      })
      .addCase(approveJoinRequest.fulfilled, (state, { payload: id }) => {
        joinRequestsAdapter.removeOne(state, id);
        state.status = 'succeeded';
      })
      ;
  },
});

export const joinRequestsReducer = joinRequestsSlice.reducer;

export default {
  joinRequests: joinRequestsReducer,
};
