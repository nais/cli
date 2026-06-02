// Based on: https://github.com/ekawahyu/printable-docs/blob/main/assets/js/theme-switch.js
// via https://github.com/just-the-docs/just-the-docs/issues/1223

window.addEventListener('DOMContentLoaded', function () {
  const themeButton = document.getElementById('theme-toggle');

  updateThemeButton(getBrowserTheme());

  jtd.addEvent(themeButton, 'click', function () {
    const theme = getBrowserTheme() === 'dark' ? 'light' : 'dark';
    setBrowserTheme(theme);
    updateThemeButton(theme);
    jtd.setTheme(theme);
    localStorage.setItem('theme', theme);
  });

  function getBrowserTheme() {
    return document.documentElement.classList.contains('dark-mode') ? 'dark' : 'light';
  }

  function setBrowserTheme(theme) {
    if (theme === 'dark') {
      document.documentElement.classList.add('dark-mode');
      document.documentElement.classList.remove('light-mode');
    } else {
      document.documentElement.classList.add('light-mode');
      document.documentElement.classList.remove('dark-mode');
    }
  }

  function updateThemeButton(theme) {
    if (theme === 'dark') {
      themeButton.innerHTML = `<svg width='18px' height='18px'><use href="#svg-moon"></use></svg>`;
    } else {
      themeButton.innerHTML = `<svg width='18px' height='18px'><use href="#svg-sun"></use></svg>`;
    }
  }
});
