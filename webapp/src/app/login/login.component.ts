import { Component, OnInit, Inject } from '@angular/core';
import { Router } from '@angular/router';
import { LoginService, ILoginService } from 'projects/ecad-angular-auth/src/public_api';


@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css']
})
export class LoginComponent implements OnInit {

  username: string;
  password: string;

  isLoggedIn: Boolean;

  constructor(
    @Inject(LoginService)
    private loginService: ILoginService,
    private router: Router,
  ) { }

  ngOnInit() {
    this.isLoggedIn = this.loginService.isLoggedIn();
  }

  async onSubmit() {
    await this.loginService.login({username: this.username, password: this.password}).toPromise();
    this.router.navigateByUrl('/');
  }
}
