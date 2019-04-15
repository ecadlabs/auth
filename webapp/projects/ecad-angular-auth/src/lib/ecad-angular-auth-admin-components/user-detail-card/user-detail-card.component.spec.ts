import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { UserDetailCardComponent } from './user-detail-card.component';
import { MatCardModule, MatDividerModule } from '@angular/material';

describe('UserDetailCardComponent', () => {
  let component: UserDetailCardComponent;
  let fixture: ComponentFixture<UserDetailCardComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [UserDetailCardComponent],
      imports: [MatCardModule, MatDividerModule]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(UserDetailCardComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
