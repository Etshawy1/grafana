import { createSelector } from '@reduxjs/toolkit';

import { selectors } from './reducers';

export const { selectAll, selectById, selectTotal: selectTotalRequests } = selectors;

const selectQuery = (_: any, query: string) => query;
export const selectJoinRequestersMatchingQuery = createSelector([selectAll, selectQuery], (joinRequests, searchQuery) => {
  const regex = new RegExp(searchQuery, 'i');
  const matches = joinRequests.filter((joinRequester) => regex.test(joinRequester.name) || regex.test(joinRequester.email));
  return matches;
});
