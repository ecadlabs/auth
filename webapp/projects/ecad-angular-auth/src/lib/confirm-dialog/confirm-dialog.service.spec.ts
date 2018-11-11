import { TestBed, inject } from '@angular/core/testing';

import { ConfirmDialogService } from './confirm-dialog.service';
import { MatDialogModule } from '@angular/material';

describe('ConfirmDialogService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [MatDialogModule],
      providers: [ConfirmDialogService]
    });
  });

  it('should be created', inject([ConfirmDialogService], (service: ConfirmDialogService) => {
    expect(service).toBeTruthy();
  }));
});
