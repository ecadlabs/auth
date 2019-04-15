import { Component, OnInit, Input, Output, EventEmitter } from '@angular/core';
import { BehaviorSubject } from 'rxjs';
import { map } from 'rxjs/operators';
import { MatTableDataSource } from '@angular/material';

@Component({
  selector: 'auth-ip-list',
  templateUrl: './ip-list.component.html',
  styleUrls: ['./ip-list.component.scss']
})
export class IpListComponent implements OnInit {
  public readonly displayedColumns = ['ip', 'action'];

  private _ips = new BehaviorSubject<string[]>([]);

  @Input()
  get ips() {
    return this._ips.value;
  }

  set ips(value: string[]) {
    this._ips.next(value);
  }

  @Output()
  remove = new EventEmitter();

  @Output()
  newIP = new EventEmitter();

  public dataSource$ = this._ips.pipe(
    map(ips => {
      return new MatTableDataSource(ips);
    })
  );

  constructor() {}

  ngOnInit() {}
}
