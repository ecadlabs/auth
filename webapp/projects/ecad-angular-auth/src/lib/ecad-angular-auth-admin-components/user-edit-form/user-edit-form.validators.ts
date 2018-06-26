import { ValidatorFn, FormControl } from '@angular/forms';

export const MinSelection = (value: number): ValidatorFn => {
    return (formControl: FormControl) => {
        if (formControl.value && Array.isArray(formControl.value) && formControl.value.length >= value) {
            return {};
        } else {
            return {
                minSelection: true
            };
        }
    };
};
