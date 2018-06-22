import { EcadAngularAuthAdminComponentsModule } from './ecad-angular-auth-admin-components.module';

describe('EcadAngularAuthAdminComponentsModule', () => {
  let ecadAngularAuthAdminComponentsModule: EcadAngularAuthAdminComponentsModule;

  beforeEach(() => {
    ecadAngularAuthAdminComponentsModule = new EcadAngularAuthAdminComponentsModule();
  });

  it('should create an instance', () => {
    expect(ecadAngularAuthAdminComponentsModule).toBeTruthy();
  });
});
