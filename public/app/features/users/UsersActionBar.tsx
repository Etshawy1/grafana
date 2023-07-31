import React from 'react';
import { connect, ConnectedProps } from 'react-redux';

import { RadioButtonGroup, LinkButton, FilterInput } from '@grafana/ui';
import config from 'app/core/config';
import { contextSrv } from 'app/core/core';
import { AccessControlAction, StoreState } from 'app/types';

import { selectTotal } from '../invites/state/selectors';
import { selectTotalRequests } from '../joinRequests/state/selectors';

import { changeSearchQuery } from './state/actions';
import { getUsersSearchQuery } from './state/selectors';

export interface OwnProps {
  showUserTypes: string;
  onShowUserTypes: (value: string) => void;
}

function mapStateToProps(state: StoreState) {
  return {
    searchQuery: getUsersSearchQuery(state.users),
    pendingInvitesCount: selectTotal(state.invites),
    externalUserMngLinkName: state.users.externalUserMngLinkName,
    externalUserMngLinkUrl: state.users.externalUserMngLinkUrl,
    canInvite: state.users.canInvite,
    joinRequestersCount: selectTotalRequests(state.joinRequests)
  };
}

const mapDispatchToProps = {
  changeSearchQuery,
};

const connector = connect(mapStateToProps, mapDispatchToProps);

export type Props = ConnectedProps<typeof connector> & OwnProps;

export const UsersActionBarUnconnected = ({
  canInvite,
  externalUserMngLinkName,
  externalUserMngLinkUrl,
  searchQuery,
  pendingInvitesCount,
  changeSearchQuery,
  onShowUserTypes,
  showUserTypes,
  joinRequestersCount
}: Props): JSX.Element => {
  const options = [
    { label: 'Users', value: 'users' },
    { label: `Pending Invites (${pendingInvitesCount})`, value: 'invites' },
    { label: `Join Requests (${joinRequestersCount})`, value: 'joinRequests' }
  ];
  const canAddToOrg: boolean = contextSrv.hasAccess(AccessControlAction.OrgUsersAdd, canInvite);
  // Show invite button in the following cases:
  // 1) the instance is not a hosted Grafana instance (!config.externalUserMngInfo)
  // 2) new basic auth users can be created for this instance (!config.disableLoginForm).
  const showInviteButton: boolean = canAddToOrg && !(config.disableLoginForm && config.externalUserMngInfo);

  return (
    <div className="page-action-bar" data-testid="users-action-bar">
      <div className="gf-form gf-form--grow">
        <FilterInput
          value={searchQuery}
          onChange={changeSearchQuery}
          placeholder="Search user by login, email or name"
        />
      </div>
      {(
        <div style={{ marginLeft: '1rem' }}>
          <RadioButtonGroup value={showUserTypes} options={options} onChange={onShowUserTypes} />
        </div>
      )}
      {showInviteButton && <LinkButton href="org/users/invite">Invite</LinkButton>}
      {externalUserMngLinkUrl && (
        <LinkButton href={externalUserMngLinkUrl} target="_blank" rel="noopener">
          {externalUserMngLinkName}
        </LinkButton>
      )}
    </div>
  );
};

export const UsersActionBar = connector(UsersActionBarUnconnected);
