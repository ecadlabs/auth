import { TestBed, inject } from '@angular/core/testing';
import { LogsService } from './logs.service';
import { HttpClientTestingModule } from '@angular/common/http/testing';

describe('LogsService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [LogsService]
    });
  });

  it('should be created', inject([LogsService], (service: LogsService) => {
    expect(service).toBeTruthy();
  }));
});
