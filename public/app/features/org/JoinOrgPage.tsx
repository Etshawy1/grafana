import React, { useEffect }  from 'react';
import { connect } from 'react-redux';
import useAsyncFn from 'react-use/lib/useAsyncFn';

import { NavModelItem } from '@grafana/data';
import { getBackendSrv, isFetchError } from '@grafana/runtime';
import { Button, RadioButtonGroup, Form, Field, InputControl, FieldSet, TextArea } from '@grafana/ui';
import { Page } from 'app/core/components/Page/Page';
import { getConfig } from 'app/core/config';
import { useAppNotification } from 'app/core/copy/appNotification';
import { GrafanaRouteComponentProps } from 'app/core/navigation/types';
import { OrgRole } from 'app/types';


const mapDispatchToProps = {
};


const roles = [
  { label: 'Viewer', value: OrgRole.Viewer },
  { label: 'Editor', value: OrgRole.Editor },
  { label: 'Admin', value: OrgRole.Admin },
];

export interface FormModel {
  role: OrgRole;
  justification: string;
}

const defaultValues: FormModel = {
  role: OrgRole.Viewer,
  justification: '',
};

const getErrorMessage = (error: Error) => {
  return isFetchError(error) ? error?.data?.message : 'An unexpected error happened.';
};

export interface JoinOrgPageRouteParams {
  orgId?: string;
}

const connector = connect(undefined, mapDispatchToProps);

export interface Props extends GrafanaRouteComponentProps<JoinOrgPageRouteParams> {}

const pageNav: NavModelItem = {
  icon: 'building',
  id: 'org-new',
  text: 'Organization Join Request',
  breadcrumbs: [{ title: 'home', url: '/' }],
};

export const NewOrgPage = ({ match, location }: Props) => {
  const notifyApp = useAppNotification();

  const [OrgAdminsState, fetchOrgAdmins] = useAsyncFn(async () =>{ return await getBackendSrv().get(`/api/user/orgAdmins/${match.params.orgId}`)}, []);
  useEffect(() => {fetchOrgAdmins();}, [fetchOrgAdmins]);

  const onSubmit = async (formData: FormModel) => {
    await getBackendSrv()
      .post(`/api/user/joinRequest/${match.params.orgId}`, formData)
      .then(() => {
        window.setTimeout(function(){
          window.location.assign(getConfig().appSubUrl + '/');
        }, 5000);
        // notifyApp.success('Join request submitted successfully');
      })
      .catch((err) => {
        const msg = err.data?.message || err;
        notifyApp.warning(msg);
      });
  };

  return (
    <Page navId="home" pageNav={pageNav}>
      <Page.Contents>
        <p className="muted">
          Do you want to submit request to join organization with id = {match.params.orgId}?
        </p>
        <p className="muted">
          Notes:
          <ol style={{marginLeft: 25}}>
            <li>Some Organizations have Viewer access join request auto approved. </li>
            <li>Normally you should reach out to organization admins to get your request approved. </li>
          </ol>
        </p>


          {OrgAdminsState.error && getErrorMessage(OrgAdminsState.error)}
          {OrgAdminsState.loading && 'Fetching Organization Admins...'}
          {OrgAdminsState.value && (
            <p className="muted">
              Organization {OrgAdminsState.value[0].name} with id={match.params.orgId} Admins Include:
              <ul style={{marginLeft: 25}}>
              {OrgAdminsState.value.map((admin: any) => (
                <li key={admin.email}>{admin.email}</li>
              ))}
              </ul>
            </p>
          )}

        <Form defaultValues={defaultValues} onSubmit={onSubmit}>
          {({ register, control, errors }) => {
            return (
              <>
                <FieldSet>
                  <Field invalid={!!errors.role} label="Role">
                    <InputControl
                      render={({ field: { ref, ...field } }) => <RadioButtonGroup {...field} options={roles} />}
                      control={control}
                      name="role"
                    />
                  </Field>
                  <Field
                    label="Justification"
                    invalid={!!errors.justification}
                    error={errors.justification?.message}
                  >
                    <TextArea
                      {...register('justification', {required: true})}
                      placeholder="why you want to join this org with the chosen role"
                      rows={3}
                    />
                  </Field>
                </FieldSet>
                <Button type="submit">Submit</Button>
              </>
            );
          }}
        </Form>
      </Page.Contents>
    </Page>
  );
};

export default connector(NewOrgPage);
