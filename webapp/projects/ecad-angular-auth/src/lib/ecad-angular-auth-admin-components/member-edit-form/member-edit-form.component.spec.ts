import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { MemberEditFormComponent } from './member-edit-form.component';
import { CommonModule } from '@angular/common';
import { MatSelectModule, MatButtonModule } from '@angular/material';
import { ReactiveFormsModule } from '@angular/forms';
import { USERS_SERVICE } from '../../ecad-angular-auth-admin/tokens';
import { Membership } from '../../ecad-angular-auth-admin/interfaces/membership.i';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';

describe('MemberEditFormComponent', () => {
  let component: MemberEditFormComponent;
  let fixture: ComponentFixture<MemberEditFormComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [MemberEditFormComponent],
      imports: [
        CommonModule,
        MatSelectModule,
        ReactiveFormsModule,
        MatButtonModule,
        NoopAnimationsModule
      ],
      providers: [
        {
          provide: USERS_SERVICE,
          useValue: { getRoles: () => [] }
        }
      ]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(MemberEditFormComponent);
    component = fixture.componentInstance;
    (component.member = { roles: {} } as Membership), fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
