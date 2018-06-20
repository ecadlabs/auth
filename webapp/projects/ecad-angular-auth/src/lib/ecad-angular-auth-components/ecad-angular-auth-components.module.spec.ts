import { EcadAngularAuthComponentsModule } from './ecad-angular-auth-components.module';

describe('EcadAngularAuthComponentsModule', () => {
  let ecadAngularAuthComponentsModule: EcadAngularAuthComponentsModule;

  beforeEach(() => {
    ecadAngularAuthComponentsModule = new EcadAngularAuthComponentsModule();
  });

  it('should create an instance', () => {
    expect(ecadAngularAuthComponentsModule).toBeTruthy();
  });
});
