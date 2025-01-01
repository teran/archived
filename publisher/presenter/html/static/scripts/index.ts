window.addEventListener('load', () => {
    const currentSwitch = localStorage.getItem('lightSwitch')?.toString()
    const htmlTag = document.getElementById('html');
    const lightSwitch = document.getElementById('lightSwitch');

    if (currentSwitch) {
        htmlTag?.setAttribute('data-bs-theme', currentSwitch);
    }

    lightSwitch?.addEventListener('click', () => {
        if (document.getElementById('html')?.getAttribute('data-bs-theme')?.toString() == "dark") {
            htmlTag?.setAttribute('data-bs-theme', 'light');
            lightSwitch?.classList.remove('dark');
            lightSwitch?.classList.add('light');
            
            localStorage.setItem('lightSwitch', 'light');
            return
        }
        htmlTag?.setAttribute('data-bs-theme', 'dark');
        lightSwitch?.classList.remove('light');
        lightSwitch?.classList.add('dark');
        
        localStorage.setItem('lightSwitch', 'dark');
        return
    });
});
