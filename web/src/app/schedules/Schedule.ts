export interface Schedule {
  id: string;
  name: string;
  everyDay: boolean;
  everyWeek: boolean;
  everyMonth: boolean;
  everyMonday: boolean;
  everyTuesday: boolean;
  everyWednesday: boolean;
  everyThurday: boolean;
  everyFriday: boolean;
  everySaturday: boolean;
  everySunday: boolean;
  dailyHour: string;
  cronSyntax: string;
}

interface Serializable<T> {
  deserialize(input: Object): T;
}
export class Schedule implements Schedule, Serializable<Schedule>{

  constructor(
  ) {  }

  deserialize(input: any) {
    this.id = input.id;
    this.name = input.name;
    return this;
  }

}
