import { Directive, Input, TemplateRef, ViewContainerRef, Inject, OnDestroy } from '@angular/core';
import { first, tap } from 'rxjs/operators';
import { LOGIN_SERVICE } from '../ecad-angular-auth/tokens';
import { ILoginService } from '../ecad-angular-auth/interfaces/login-service.i';

import { distinctUntilChanged } from 'rxjs/operators';

@Directive({
    selector: '[authEcadPermissions]'
})
export class EcadPermissionsDirective implements OnDestroy {

    private permissionsSub = null;

    @Input() set authEcadPermissions(permissions: string) {
        this.permissionsSub = this.loginService.hasPermissions(permissions.split(','))
            .pipe(
                distinctUntilChanged(),
                tap((hasPermissions) => {
                    if (hasPermissions && this.templateRef) {
                        this.viewContainer.createEmbeddedView(this.templateRef);
                    } else {
                        this.viewContainer.clear();
                    }
                })
            ).subscribe();
    }

    ngOnDestroy() {
        if (this.permissionsSub) {
            this.permissionsSub.unsubscribe();
        }
    }

    constructor(
        private templateRef: TemplateRef<any>,
        private viewContainer: ViewContainerRef,
        @Inject(LOGIN_SERVICE)
        private loginService: ILoginService
    ) { }
}
