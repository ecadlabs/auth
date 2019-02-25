import {
    EcadAngularAuthModule,
    EcadAngularAuthComponentsModule,
    EcadAngularAuthAdminComponentsModule,
    EcadAngularAuthAdminModule
} from 'projects/ecad-angular-auth/src/public_api';

import { tokenGetter, tokenSetter } from 'src/app/app.module';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { RouterTestingModule } from '@angular/router/testing';

export const ecadAngularAuth = [
    NoopAnimationsModule,
    RouterTestingModule,
    EcadAngularAuthModule.forRoot({
        loginUrl: '/api/v1/login',
        whiteListUrl: '/api/v1/checkip',
        tokenGetter,
        tokenSetter,
        passwordResetUrl: '/api/v1/password_reset',
        sendResetEmailUrl: '/api/v1/request_password_reset',
        loginPageUrl: '',
        roleGuardRedirectUrl: '',
        autoRefreshInterval: 5000,
        tokenPropertyPrefix: 'com.ecadlabs.auth',
        rolesPermissionsMapping: {
            'com.ecadlabs.auth.admin': ['show.is-admin']
        },
        emailChangeValidationUrl: '/api/v1/email_update',
        emailUpdateUrl: '/api/v1/request_email_update'
    }),
    EcadAngularAuthComponentsModule,
    EcadAngularAuthAdminComponentsModule,
    EcadAngularAuthAdminModule.forRoot({
        roles: [
            { value: 'com.ecadlabs.auth.regular', displayValue: 'Regular' },
            { value: 'com.ecadlabs.auth.admin', displayValue: 'Admin' }
        ],
        apiEndpoint: '/api/v1/users',
        emailUpdateUrl: '/api/v1/request_email_update'
    })
];
