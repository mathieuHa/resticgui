export interface Schedule {
  id: string;
  every_day: boolean;
  every_week: boolean;
  every_month: boolean;
  every_monday: boolean;
  every_tuesday: boolean;
  every_wednesday: boolean;
  every_thursday: boolean;
  every_friday: boolean;
  every_saturday: boolean;
  every_sunday: boolean;
  hour: string;
  day: string;
  cron_syntax: string;
  interval: string;
}

interface Serializable<T> {
  deserialize(input: Object): T;
}
export class Schedule implements Schedule, Serializable<Schedule>{

  constructor(
  ) {  }

  deserialize(input: any) {
    this.id = input.id;
    this.day = input.day;
    this.every_monday = input.every_monday;
    this.every_tuesday = input.every_tuesday;
    this.every_wednesday = input.every_wednesday;
    this.every_thursday = input.every_thursday;
    this.every_friday = input.every_friday;
    this.every_saturday = input.every_saturday;
    this.every_sunday = input.every_sunday;
    this.every_week = input.every_week;
    this.every_day = input.every_day;
    this.every_month = input.every_month;
    this.interval = input.interval;
    this.cron_syntax = input.cron_syntax;
    return this;
  }

}
