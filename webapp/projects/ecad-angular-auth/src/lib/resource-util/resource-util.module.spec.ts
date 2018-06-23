import { ResourceUtilModule } from './resource-util.module';

describe('ResourceUtilModule', () => {
  let resourceUtilModule: ResourceUtilModule;

  beforeEach(() => {
    resourceUtilModule = new ResourceUtilModule();
  });

  it('should create an instance', () => {
    expect(resourceUtilModule).toBeTruthy();
  });
});
