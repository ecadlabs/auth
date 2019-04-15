import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { IpListComponent } from './ip-list.component';
import {
  MatButtonModule,
  MatSortModule,
  MatIconModule,
  MatTableModule,
  MatTooltipModule
} from '@angular/material';

describe('IpListComponent', () => {
  let component: IpListComponent;
  let fixture: ComponentFixture<IpListComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [IpListComponent],
      imports: [
        MatButtonModule,
        MatSortModule,
        MatIconModule,
        MatTableModule,
        MatTooltipModule
      ]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(IpListComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
