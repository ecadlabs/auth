import { Component, Inject } from '@angular/core';
import { LOGIN_SERVICE, ILoginService } from 'ecad-angular-auth';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent {
  title = 'app';

  constructor(
    @Inject(LOGIN_SERVICE)
    private loginService: ILoginService,
  ) {

  }

  isLoggedIn = this.loginService.isLoggedIn;
  user = this.loginService.user;

  logout() {
    this.loginService.logout().subscribe();
  }
}
