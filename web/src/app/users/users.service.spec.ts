import { TestBed } from '@angular/core/testing';

import { UsersService } from './users.service';
import {HttpClient, HttpClientModule, HttpErrorResponse} from "@angular/common/http";
import {User} from "./user";
import {HttpClientTestingModule, HttpTestingController} from "@angular/common/http/testing";
import {Data} from "@angular/router";

describe('UsersService', () => {
  let service: UsersService;
  describe('HttpClient testing', () => {
    let httpClient: HttpClient;
    let httpTestingController: HttpTestingController;
    beforeEach(() => {

      TestBed.configureTestingModule({
        imports: [HttpClientTestingModule]
      });
      httpClient = TestBed.inject(HttpClient);
      httpTestingController = TestBed.inject(HttpTestingController);
      //service = TestBed.inject(UsersService);
    });

    afterEach(() => {
      // After every test, assert that there are no more pending requests.
      httpTestingController.verify();
    });

    it('can test HttpClient.getUsers', () => {
      const testData: Data = [
        {name: 'Test Data', id: 1},
        {name: 'Test Data 2', id: 2}
      ];
      // Make an HTTP GET request
      httpClient.get<Data>("/api/users")
        .subscribe(data =>
          // When observable resolves, result should match test data
          expect(data).toEqual(testData)
        );
      // The following `expectOne()` will match the request's URL.
      // If no requests or multiple requests matched that URL
      // `expectOne()` would throw.
      const req = httpTestingController.expectOne('/api/users');
      // Assert that the request is a GET.
      expect(req.request.method).toEqual('GET');
      // Respond with mock data, causing Observable to resolve.
      // Subscribe callback asserts that correct data was returned.
      req.flush(testData);
    });

    it('can test HttpClient.getUsers One', () => {
      const testData: Data = {name: 'Test Data', id: 1};
      // Make an HTTP GET request
      httpClient.get<Data>("/api/users/1")
        .subscribe(data =>
          // When observable resolves, result should match test data
          expect(data).toEqual(testData)
        );
      // The following `expectOne()` will match the request's URL.
      // If no requests or multiple requests matched that URL
      // `expectOne()` would throw.
      const req = httpTestingController.expectOne('/api/users/1');
      // Assert that the request is a GET.
      expect(req.request.method).toEqual('GET');
      // Respond with mock data, causing Observable to resolve.
      // Subscribe callback asserts that correct data was returned.
      req.flush(testData);
    });
  })

  it('should be created', () => {
    TestBed.configureTestingModule({
      imports: [HttpClientModule]
    });
    service = TestBed.inject(UsersService);
    expect(service).toBeTruthy();
  });

  it('should use api/users', () => {
    TestBed.configureTestingModule({
      imports: [HttpClientModule]
    });
    service = TestBed.inject(UsersService);
    expect(service.userUrl).toBe('api/users');
  });


});
