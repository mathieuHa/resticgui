export interface Schedule {
  id: string;
  Name: string;
  EveryDay: boolean;
  EveryWeek: boolean;
  EveryMonth: boolean;
  EveryMonday: boolean;
  EveryTuesday: boolean;
  EveryWednesday: boolean;
  EveryThursday: boolean;
  EveryFriday: boolean;
  EverySaturday: boolean;
  EverySunday: boolean;
  Hour: string;
  Day: string;
  CronSyntax: string;
  Period: string;
}

interface Serializable<T> {
  deserialize(input: Object): T;
}
export class Schedule implements Schedule, Serializable<Schedule>{

  constructor(
  ) {  }

  deserialize(input: any) {
    this.id = input.id;
    this.Name = input.Name;
    this.Day = input.Day;
    this.EveryMonday = input.EveryMonday;
    this.EveryTuesday = input.EveryTuesday;
    this.EveryWednesday = input.EveryWednesday;
    this.EveryThursday = input.EveryThursday;
    this.EveryFriday = input.EveryFriday;
    this.EverySaturday = input.EverySaturday;
    this.EverySunday = input.EverySunday;
    this.EveryWeek = input.EveryWeek;
    this.EveryDay = input.EveryDay;
    this.EveryMonth = input.EveryMonth;
    this.Period = input.Period;
    this.CronSyntax = input.CronSyntax;
    return this;
  }

}
