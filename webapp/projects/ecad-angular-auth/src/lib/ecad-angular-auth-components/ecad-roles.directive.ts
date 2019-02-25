import { Directive, Input, TemplateRef, ViewContainerRef, Inject, OnDestroy } from '@angular/core';
import { tap, distinctUntilChanged } from 'rxjs/operators';
import { LOGIN_SERVICE } from '../ecad-angular-auth/tokens';
import { ILoginService } from '../ecad-angular-auth/interfaces/login-service.i';

@Directive({
  selector: '[authEcadRoles]'
})
export class EcadRolesDirective implements OnDestroy {
  private rolesSub = null;

  @Input() set authEcadRoles(allowedRoles: string) {
    allowedRoles = allowedRoles.replace(/\s/g,'');

    this.rolesSub = this.loginService.hasOneOfRoles(allowedRoles.split(',')).pipe(
      distinctUntilChanged(),
      tap(hasRole => {
        if (hasRole && this.templateRef) {
          this.viewContainer.createEmbeddedView(this.templateRef);
        } else {
          this.viewContainer.clear();
        }
      })
    ).subscribe();
  }

  ngOnDestroy() {
    if (this.rolesSub) {
      this.rolesSub.unsubscribe();
    }
  }

  constructor(
    private templateRef: TemplateRef<any>,
    private viewContainer: ViewContainerRef,
    @Inject(LOGIN_SERVICE)
    private loginService: ILoginService
  ) {}
}
