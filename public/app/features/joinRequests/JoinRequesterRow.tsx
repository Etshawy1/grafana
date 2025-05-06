import React, { PureComponent } from 'react';
import { connect, ConnectedProps } from 'react-redux';

import { Button } from '@grafana/ui';
import { JoinRequester } from 'app/types';

import { rejectJoinRequest, approveJoinRequest } from './state/actions';

const mapDispatchToProps = {
  rejectJoinRequest,
  approveJoinRequest
};

const connector = connect(null, mapDispatchToProps);

interface OwnProps {
  joinRequester: JoinRequester;
}

export type Props = OwnProps & ConnectedProps<typeof connector>;

class JoinRequesterRow extends PureComponent<Props> {
  render() {
    const { joinRequester, rejectJoinRequest, approveJoinRequest } = this.props;
    return (
      <tr>

        <td className="max-width-4">
          <span className="ellipsis" title={joinRequester.email}>
                    {joinRequester.email}
          </span>
        </td>
        <td className="width-3">
          <span className="ellipsis" title={joinRequester.role}>
                    {joinRequester.role}
          </span>
        </td>
        <td className="max-width-9">
          <span className="ellipsis" title={joinRequester.justification}>
                    {joinRequester.justification}
          </span>
        </td>
        <td className="width-1 text-right" style={{padding:1}}>
        <Button
            variant="success"
            size="sm"
            icon="check"
            onClick={() => approveJoinRequest(joinRequester.id)}
            aria-label="Accept Join Request"
          />
        </td>
        <td className="width-1 text-right" style={{padding:1}}>
          <Button
            variant="destructive"
            size="sm"
            icon="times"
            onClick={() => rejectJoinRequest(joinRequester.id)}
            aria-label="Reject Join Request"
          />
        </td>
      </tr>
    );
  }
}

export default connector(JoinRequesterRow);
