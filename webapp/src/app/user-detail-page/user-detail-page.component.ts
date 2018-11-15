import { Component, OnInit, ChangeDetectorRef } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { pluck } from 'rxjs/operators';
import { User } from '@ecadlabs/angular-auth';

@Component({
  selector: 'app-user-detail-page',
  templateUrl: './user-detail-page.component.html',
  styleUrls: ['./user-detail-page.component.scss']
})
export class UserDetailPageComponent implements OnInit {

  public userId;

  public userId$ = this.activatedRoute.params.pipe(
    pluck('id'),
  );

  constructor(
    private activatedRoute: ActivatedRoute,
    private router: Router,
    private changeDetector: ChangeDetectorRef
  ) { }

  userClicked($event: User) {
    this.router.navigateByUrl(`/user/${$event.id}`);
  }

  ngOnInit() {
  }

}
