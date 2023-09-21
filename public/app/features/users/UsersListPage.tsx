import React, { useEffect, useState } from 'react';
import { connect, ConnectedProps } from 'react-redux';
import { useAsyncFn } from 'react-use';

import { renderMarkdown } from '@grafana/data';
import { selectors as e2eSelectors } from '@grafana/e2e-selectors/src';
import { getBackendSrv } from '@grafana/runtime';
import { HorizontalGroup, Pagination, VerticalGroup } from '@grafana/ui';
import { Page } from 'app/core/components/Page/Page';
import { contextSrv } from 'app/core/core';
import { OrgUser, OrgRole, StoreState } from 'app/types';

import InviteesTable from '../invites/InviteesTable';
import { fetchInvitees } from '../invites/state/actions';
import { selectInvitesMatchingQuery } from '../invites/state/selectors';
import JoinRequestersTable from '../joinRequests/JoinRequestersTable';
import { fetchJoinRequesters } from '../joinRequests/state/actions';
import { selectJoinRequestersMatchingQuery } from '../joinRequests/state/selectors';

import { UsersActionBar } from './UsersActionBar';
import { UsersTable } from './UsersTable';
import { loadUsers, removeUser, updateUser, changePage } from './state/actions';
import { getUsers, getUsersSearchQuery } from './state/selectors';

function mapStateToProps(state: StoreState) {
  const searchQuery = getUsersSearchQuery(state.users);
  return {
    users: getUsers(state.users),
    searchQuery: getUsersSearchQuery(state.users),
    page: state.users.page,
    totalPages: state.users.totalPages,
    perPage: state.users.perPage,
    invitees: selectInvitesMatchingQuery(state.invites, searchQuery),
    joinRequesters: selectJoinRequestersMatchingQuery(state.joinRequests, searchQuery),
    externalUserMngInfo: state.users.externalUserMngInfo,
    isLoading: state.users.isLoading,
  };
}

const mapDispatchToProps = {
  loadUsers,
  fetchInvitees,
  fetchJoinRequesters,
  changePage,
  updateUser,
  removeUser,
};

const connector = connect(mapStateToProps, mapDispatchToProps);

export type Props = ConnectedProps<typeof connector>;

export interface State {
  showUserTypes: string;
}

const selectors = e2eSelectors.pages.UserListPage.UsersListPage;

export const UsersListPageUnconnected = ({
  users,
  page,
  totalPages,
  invitees,
  joinRequesters,
  externalUserMngInfo,
  isLoading,
  loadUsers,
  fetchInvitees,
  fetchJoinRequesters,
  changePage,
  updateUser,
  removeUser,
}: Props): JSX.Element => {
  const [showUserTypes, setshowUserTypes] = useState("users");
  const [activeOrg, fetchOrg] = useAsyncFn(async () =>{ return await getBackendSrv().get(`/api/orgs/${contextSrv.user.orgId}`)}, []);
  const externalUserMngInfoHtml = externalUserMngInfo ? renderMarkdown(externalUserMngInfo) : '';

  useEffect(() => {
    loadUsers();
    fetchInvitees();
    fetchJoinRequesters();
    fetchOrg();
  }, [fetchJoinRequesters, fetchInvitees, loadUsers, fetchOrg]);

  const onRoleChange = (role: OrgRole, user: OrgUser) => {
    updateUser({ ...user, role: role });
  };

  const onShowUserTypes = (value: string) => {
    setshowUserTypes(value);
    loadUsers();
    fetchInvitees();
    fetchJoinRequesters();
  };

  const onSwitchAutoApproveJoinRequests = () => {
    getBackendSrv().put(`/api/orgs/${contextSrv.user.orgId}/autoApprove`, {autoApprove: !activeOrg.value.autoApproveJoinRequests})
    .then(() => {
      fetchOrg();
    });
  }

  const renderTable = () => {
    if (showUserTypes === 'invites') {
      return <InviteesTable invitees={invitees} />;
    } else if(showUserTypes === 'joinRequests') {
      return <JoinRequestersTable joinRequesters={joinRequesters} org={activeOrg} onSwitch={onSwitchAutoApproveJoinRequests}/>
    } else {
      return (
        <VerticalGroup spacing="md" data-testid={selectors.container}>
          <UsersTable
            users={users}
            orgId={contextSrv.user.orgId}
            onRoleChange={(role, user) => onRoleChange(role, user)}
            onRemoveUser={(user) => removeUser(user.userId)}
          />
          <HorizontalGroup justify="flex-end">
            <Pagination
              onNavigate={changePage}
              currentPage={page}
              numberOfPages={totalPages}
              hideWhenSinglePage={true}
            />
          </HorizontalGroup>
        </VerticalGroup>
      );
    }
  };

  return (
    <Page.Contents isLoading={!isLoading}>
      <UsersActionBar onShowUserTypes={onShowUserTypes} showUserTypes={showUserTypes} />
      {externalUserMngInfoHtml && (
        <div className="grafana-info-box" dangerouslySetInnerHTML={{ __html: externalUserMngInfoHtml }} />
      )}
      {isLoading && renderTable()}
    </Page.Contents>
  );
};

export const UsersListPageContent = connector(UsersListPageUnconnected);

export default function UsersListPage() {
  return (
    <Page navId="users">
      <UsersListPageContent />
    </Page>
  );
}
