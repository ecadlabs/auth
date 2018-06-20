import { Component, Inject } from '@angular/core';
import { LoginService, ILoginService } from 'ecad-angular-auth';
import { map } from 'rxjs/operators';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent {
  title = 'app';

  constructor(
    @Inject(LoginService)
    private loginService: ILoginService,
  ) {

  }

  isLoggedIn = this.loginService.isLoggedIn;
  user = this.loginService.user;

  logout() {
    this.loginService.logout().subscribe();
  }
}
