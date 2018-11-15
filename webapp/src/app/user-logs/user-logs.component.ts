import { Component, OnInit, ChangeDetectorRef } from '@angular/core';
import { User } from '@ecadlabs/angular-auth';
import { Router, ActivatedRoute } from '@angular/router';

@Component({
  selector: 'app-user-logs',
  templateUrl: './user-logs.component.html',
  styleUrls: ['./user-logs.component.scss']
})
export class UserLogsComponent implements OnInit {

  constructor(
    private activatedRoute: ActivatedRoute,
    private router: Router,
    private changeDetector: ChangeDetectorRef
  ) { }

  ngOnInit() {
  }

  userClicked($event: User) {
    this.router.navigateByUrl(`/user/${$event.id}`);
  }


}
