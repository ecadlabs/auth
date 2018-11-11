import { TestBed, inject } from '@angular/core/testing';

import { ResourcesService } from './resources.service';
import { HttpClientTestingModule } from '@angular/common/http/testing';

describe('ResourcesService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [ResourcesService]
    });
  });

  it('should be created', inject([ResourcesService], (service: ResourcesService<any, any>) => {
    expect(service).toBeTruthy();
  }));
});
