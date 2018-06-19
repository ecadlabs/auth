import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';

import { RouterModule } from '@angular/router';
import { MatInputModule, MatCardModule, MatButtonModule } from '@angular/material';
import { AppComponent } from './app.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { LoginComponent } from './login/login.component';
import { ReactiveFormsModule, FormsModule } from '@angular/forms';
import { EcadAngularAuthModule } from 'ecad-angular-auth';



@NgModule({
  declarations: [
    AppComponent,
    LoginComponent
  ],
  imports: [
    EcadAngularAuthModule.forRoot({
      loginUrl: 'test',
      whiteListUrl: 'test',
      tokenName: 'test',
      passwordResetUrl: 'test',
    }),
    BrowserModule,
    BrowserAnimationsModule,
    RouterModule.forRoot([
      {path: '', pathMatch: 'full', component: LoginComponent}
    ]),
    FormsModule,
    MatInputModule,
    MatCardModule,
    MatButtonModule
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
