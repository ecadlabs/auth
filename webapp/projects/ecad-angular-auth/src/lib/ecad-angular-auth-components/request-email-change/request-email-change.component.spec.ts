import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { RequestEmailChangeComponent } from './request-email-change.component';
import { MatCardModule, MatInputModule, MatButtonModule } from '@angular/material';
import { LOGIN_SERVICE, AUTH_CONFIG } from '../../ecad-angular-auth/tokens';
import { RouterTestingModule } from '@angular/router/testing';
import { ReactiveFormsModule } from '@angular/forms';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';

describe('RequestEmailChangeComponent', () => {
  let component: RequestEmailChangeComponent;
  let fixture: ComponentFixture<RequestEmailChangeComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        MatCardModule,
        MatInputModule,
        MatButtonModule,
        RouterTestingModule,
        ReactiveFormsModule,
        NoopAnimationsModule
      ],
      providers: [
        { provide: LOGIN_SERVICE, useValue: {} },
        { provide: AUTH_CONFIG, useValue: {} }
      ],
      declarations: [RequestEmailChangeComponent],
      schemas: [CUSTOM_ELEMENTS_SCHEMA]
    })
      .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(RequestEmailChangeComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
