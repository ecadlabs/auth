import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ResetPasswordEmailComponent } from './reset-password-email.component';
import { ecadAngularAuth } from 'src/testing/fixture';

describe('ResetPasswordEmailComponent', () => {
  let component: ResetPasswordEmailComponent;
  let fixture: ComponentFixture<ResetPasswordEmailComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        ...ecadAngularAuth
      ],
      declarations: [ResetPasswordEmailComponent]
    })
      .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ResetPasswordEmailComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
