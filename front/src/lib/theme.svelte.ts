export type ThemeChoice = 'system' | 'light' | 'dark';

const KEY = 'resticgui-theme';

function read(): ThemeChoice {
  try {
    const v = localStorage.getItem(KEY);
    return v === 'light' || v === 'dark' ? v : 'system';
  } catch {
    return 'system';
  }
}

/** The theme the user picked; "system" follows the operating system. */
class Theme {
  choice = $state<ThemeChoice>(read());

  set(choice: ThemeChoice) {
    this.choice = choice;
    try {
      if (choice === 'system') localStorage.removeItem(KEY);
      else localStorage.setItem(KEY, choice);
    } catch {
      // private mode: the choice lasts until the page is closed
    }
    apply(choice);
  }
}

/** Sets data-theme on <html>; app.html does the same before the first paint. */
export function apply(choice: ThemeChoice, root = document.documentElement) {
  if (choice === 'system') delete root.dataset.theme;
  else root.dataset.theme = `resticgui-${choice}`;
}

export const theme = new Theme();
