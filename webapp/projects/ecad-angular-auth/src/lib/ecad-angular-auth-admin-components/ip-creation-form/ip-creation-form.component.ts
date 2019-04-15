import { Component, OnInit, Output, EventEmitter } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';

@Component({
  selector: 'auth-ip-creation-form',
  templateUrl: './ip-creation-form.component.html',
  styleUrls: ['./ip-creation-form.component.scss']
})
export class IpCreationFormComponent implements OnInit {
  private readonly CDIRRegex = /^([0-9]{1,3}\.){3}[0-9]{1,3}(\/([0-9]|[1-2][0-9]|3[0-2]))?$/;
  public ipForm: FormGroup;

  @Output()
  newIp = new EventEmitter();

  constructor(private _fb: FormBuilder) {
    this.ipForm = this._fb.group({
      ip: ['', [Validators.required, Validators.pattern(this.CDIRRegex)]]
    });
  }

  ngOnInit() {}

  public submit() {
    this.newIp.next(this.ipForm.value.ip);
  }
}
