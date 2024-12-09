import React, { PureComponent } from 'react';

import { Switch, Field } from '@grafana/ui';
import { JoinRequester } from 'app/types';

import JoinRequesterRow from './JoinRequesterRow';

export interface Props {
  joinRequesters: JoinRequester[];
  org: any;
  onSwitch: any;
}

export default class JoinRequestersTable extends PureComponent<Props> {
  render() {
    const { joinRequesters, org, onSwitch } = this.props;

    return (
      <>
        {org.loading && 'Fetching Organization Auto Approval State...'}
        {org.value && (
          <Field label="Auto Approve 'Viewer' Role Join Requests to this organization">
            <Switch
              id="auto-approve-join-requests"
              value={!!org.value.autoApproveJoinRequests}
              onChange={onSwitch}
              />
          </Field>
        )}
        <table className="filter-table form-inline">
          <thead>
            <tr>
              <th>Email</th>
              <th>Role</th>
              <th>Justification</th>
              <th />
              <th style={{ width: '34px' }} />
            </tr>
          </thead>
          <tbody>
            {joinRequesters.map((joinRequester, index) => {
              return <JoinRequesterRow key={`${joinRequester.id}-${index}`} joinRequester={joinRequester} />;
            })}
          </tbody>
        </table>
      </>
    );
  }
}
