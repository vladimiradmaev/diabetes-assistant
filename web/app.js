// Основной JavaScript файл приложения "Помощник диабетика"

// Generate a unique user ID if not already exists
function generateUserId() {
    const userId = 'user_' + Math.random().toString(36).substring(2, 15);
    localStorage.setItem('diabetes_user_id', userId);
    return userId;
}

let currentUserId = localStorage.getItem('diabetes_user_id') || generateUserId();
let bloodSugarChart = null;
let userSettings = null;

// Define API base URL
const API_BASE_URL = 'http://localhost:8080/api';

// Инициализация приложения при загрузке DOM
document.addEventListener('DOMContentLoaded', function() {
    // Отображение ID пользователя
    const userIdDisplay = document.getElementById('user-id-display');
    if (userIdDisplay) {
        userIdDisplay.textContent = currentUserId;
    }
    
    // Настройка навигации
    setupNavigation();
    
    // Настройка обработчиков событий для форм
    setupEventListeners();
    
    // Показ главной страницы по умолчанию
    showPage('dashboard');
    
    // Загружаем данные
    loadUserSettings();
    loadBloodSugarReadings();

    // Setup initial insulin periods
    setupInitialInsulinPeriods();
});

function setupNavigation() {
    document.querySelectorAll('.nav-link').forEach(link => {
        link.addEventListener('click', function(e) {
            e.preventDefault();
            const href = this.getAttribute('href');
            if (href && href.startsWith('#')) {
                const pageId = href.substring(1); // Удаляем '#' из href
                showPage(pageId);
            }
        });
    });
}

function setupEventListeners() {
    // Настраиваем обработчики событий для форм
    const settingsForm = document.getElementById('settings-form');
    if (settingsForm) {
        settingsForm.addEventListener('submit', function(e) {
            e.preventDefault();
            saveSettings(e);
        });
    }
    
    // Настройка кнопок для управления периодами инсулина
    const addPeriodBtn = document.getElementById('add-period-btn');
    if (addPeriodBtn) {
        addPeriodBtn.addEventListener('click', addInsulinPeriod);
        
        // Устанавливаем начальные периоды при загрузке страницы
        setupInitialInsulinPeriods();
        
        // Добавляем обработчик для отслеживания изменений в периодах
        const coefficientsContainer = document.getElementById('insulin-coefficients-container');
        if (coefficientsContainer) {
            coefficientsContainer.addEventListener('change', function(e) {
                if (e.target.classList.contains('period-start') || 
                    e.target.classList.contains('period-end')) {
                    validatePeriods('insulin');
                }
            });
            
            // Добавляем делегирование событий для кнопок удаления
            coefficientsContainer.addEventListener('click', function(e) {
                if (e.target.classList.contains('remove-period')) {
                    removePeriod(e, 'insulin');
                }
            });
        }
    }
    
    // Настройка кнопок для управления периодами чувствительности к инсулину
    const addSensitivityBtn = document.getElementById('add-sensitivity-period-btn');
    if (addSensitivityBtn) {
        addSensitivityBtn.addEventListener('click', addSensitivityPeriod);
        
        // Добавляем обработчик для отслеживания изменений в периодах
        const sensitivityContainer = document.getElementById('insulin-sensitivity-container');
        if (sensitivityContainer) {
            sensitivityContainer.addEventListener('change', function(e) {
                if (e.target.classList.contains('sensitivity-start') || 
                    e.target.classList.contains('sensitivity-end')) {
                    validatePeriods('sensitivity');
                }
            });
            
            // Добавляем делегирование событий для кнопок удаления
            sensitivityContainer.addEventListener('click', function(e) {
                if (e.target.classList.contains('remove-period')) {
                    removePeriod(e, 'sensitivity');
                }
            });
        }
    }
    
    // Настройка кнопок для управления периодами соотношения инсулина к углеводам
    const addCarbRatioBtn = document.getElementById('add-carb-ratio-period-btn');
    if (addCarbRatioBtn) {
        addCarbRatioBtn.addEventListener('click', addCarbRatioPeriod);
        
        // Добавляем обработчик для отслеживания изменений в периодах
        const carbRatioContainer = document.getElementById('carb-ratio-container');
        if (carbRatioContainer) {
            carbRatioContainer.addEventListener('change', function(e) {
                if (e.target.classList.contains('carb-ratio-start') || 
                    e.target.classList.contains('carb-ratio-end')) {
                    validatePeriods('carb-ratio');
                }
            });
            
            // Добавляем делегирование событий для кнопок удаления
            carbRatioContainer.addEventListener('click', function(e) {
                if (e.target.classList.contains('remove-period')) {
                    removePeriod(e, 'carb-ratio');
                }
            });
        }
    }
    
    const bloodSugarForm = document.getElementById('blood-sugar-form');
    if (bloodSugarForm) {
        bloodSugarForm.addEventListener('submit', function(e) {
    e.preventDefault();
            saveBloodSugar('blood-sugar-value');
        });
    }
    
    const bloodSugarFormFull = document.getElementById('blood-sugar-form-full');
    if (bloodSugarFormFull) {
        bloodSugarFormFull.addEventListener('submit', function(e) {
            e.preventDefault();
            saveBloodSugar('blood-sugar-value-full');
        });
    }
    
    const foodAnalysisForm = document.getElementById('food-analysis-form');
    if (foodAnalysisForm) {
        foodAnalysisForm.addEventListener('submit', analyzeFood);
    }
    
    const syncLibreBtn = document.getElementById('sync-libre-btn');
    if (syncLibreBtn) {
        syncLibreBtn.addEventListener('click', syncWithLibre);
    }
}

// Функция для настройки начальных периодов инсулина
function setupInitialInsulinPeriods() {
    // Clear existing periods for base insulin coefficients
    const insulinContainer = document.getElementById('insulin-coefficients-container');
    insulinContainer.innerHTML = '';
    
    // Add default periods for base insulin coefficients
    addPeriodToContainer(insulinContainer, { startHour: 0, endHour: 6, name: 'Ночь', coefficient: 1.0 }, 'insulin');
    addPeriodToContainer(insulinContainer, { startHour: 6, endHour: 12, name: 'Утро', coefficient: 1.2 }, 'insulin');
    addPeriodToContainer(insulinContainer, { startHour: 12, endHour: 18, name: 'День', coefficient: 1.0 }, 'insulin');
    addPeriodToContainer(insulinContainer, { startHour: 18, endHour: 24, name: 'Вечер', coefficient: 1.1 }, 'insulin');
    
    // Clear existing periods for insulin sensitivity
    const sensitivityContainer = document.getElementById('insulin-sensitivity-container');
    sensitivityContainer.innerHTML = '';
    
    // Add default periods for insulin sensitivity
    addPeriodToContainer(sensitivityContainer, { startHour: 0, endHour: 6, name: 'Ночь', coefficient: 1.8 }, 'sensitivity');
    addPeriodToContainer(sensitivityContainer, { startHour: 6, endHour: 12, name: 'Утро', coefficient: 1.5 }, 'sensitivity');
    addPeriodToContainer(sensitivityContainer, { startHour: 12, endHour: 18, name: 'День', coefficient: 1.5 }, 'sensitivity');
    addPeriodToContainer(sensitivityContainer, { startHour: 18, endHour: 24, name: 'Вечер', coefficient: 1.7 }, 'sensitivity');
    
    // Clear existing periods for carb ratio
    const carbRatioContainer = document.getElementById('carb-ratio-container');
    carbRatioContainer.innerHTML = '';
    
    // Add default periods for carb ratio
    addPeriodToContainer(carbRatioContainer, { startHour: 0, endHour: 6, name: 'Ночь', coefficient: 12.0 }, 'carb-ratio');
    addPeriodToContainer(carbRatioContainer, { startHour: 6, endHour: 12, name: 'Утро', coefficient: 10.0 }, 'carb-ratio');
    addPeriodToContainer(carbRatioContainer, { startHour: 12, endHour: 18, name: 'День', coefficient: 10.0 }, 'carb-ratio');
    addPeriodToContainer(carbRatioContainer, { startHour: 18, endHour: 24, name: 'Вечер', coefficient: 11.0 }, 'carb-ratio');
    
    // Validate all period types
    validatePeriods('insulin');
    validatePeriods('sensitivity');
    validatePeriods('carb-ratio');
}

function addPeriodToContainer(container, period, type) {
    // Создание элемента строки
    const row = document.createElement('div');
    row.className = 'period-row row mb-2';
    
    // HTML для содержимого строки, зависящее от типа периода
    let html = '';
    
    if (type === 'insulin') {
        html = `
            <div class="col-md-3">
                <input type="text" class="form-control period-name" value="${period.name || ''}" placeholder="Название">
            </div>
            <div class="col-md-3">
                <input type="time" class="form-control period-start" value="${formatTime(period.startHour)}" placeholder="Начало периода">
            </div>
            <div class="col-md-3">
                <input type="time" class="form-control period-end" value="${formatTime(period.endHour)}" placeholder="Конец периода">
            </div>
            <div class="col-md-2">
                <input type="number" class="form-control period-coefficient" min="0.1" step="0.1" value="${period.coefficient !== undefined ? period.coefficient : 1.0}" placeholder="Коэффициент">
            </div>
            <div class="col-md-1">
                <button type="button" class="btn btn-danger btn-sm remove-period">Удалить</button>
            </div>
        `;
    } else if (type === 'sensitivity') {
        html = `
            <div class="col-md-3">
                <input type="text" class="form-control sensitivity-name" value="${period.name || ''}" placeholder="Название">
            </div>
            <div class="col-md-3">
                <input type="time" class="form-control sensitivity-start" value="${formatTime(period.startHour)}" placeholder="Начало периода">
            </div>
            <div class="col-md-3">
                <input type="time" class="form-control sensitivity-end" value="${formatTime(period.endHour)}" placeholder="Конец периода">
            </div>
            <div class="col-md-2">
                <input type="number" class="form-control sensitivity-coefficient" min="0.1" step="0.1" value="${period.coefficient !== undefined ? period.coefficient : 2.0}" placeholder="ммоль/л на ед.">
            </div>
            <div class="col-md-1">
                <button type="button" class="btn btn-danger btn-sm remove-period">Удалить</button>
            </div>
        `;
    } else if (type === 'carb-ratio') {
        html = `
            <div class="col-md-3">
                <input type="text" class="form-control carb-ratio-name" value="${period.name || ''}" placeholder="Название">
            </div>
            <div class="col-md-3">
                <input type="time" class="form-control carb-ratio-start" value="${formatTime(period.startHour)}" placeholder="Начало периода">
            </div>
            <div class="col-md-3">
                <input type="time" class="form-control carb-ratio-end" value="${formatTime(period.endHour)}" placeholder="Конец периода">
            </div>
            <div class="col-md-2">
                <input type="number" class="form-control carb-ratio-coefficient" min="0.1" step="0.1" value="${period.coefficient !== undefined ? period.coefficient : 10.0}" placeholder="г на ед.">
            </div>
            <div class="col-md-1">
                <button type="button" class="btn btn-danger btn-sm remove-period">Удалить</button>
            </div>
        `;
    }
    
    row.innerHTML = html;
    container.appendChild(row);
    
    // Добавляем обработчик для кнопки удаления
    const removeButton = row.querySelector('.remove-period');
    if (removeButton) {
        removeButton.addEventListener('click', function() {
            container.removeChild(row);
            validatePeriods(type);
        });
    }
    
    // Добавляем обработчики для всех полей ввода
    const inputs = row.querySelectorAll('input');
    inputs.forEach(input => {
        input.addEventListener('change', function() {
            validatePeriods(type);
        });
    });
    
    // Validate immediately after adding
    validatePeriods(type);
}

function formatTime(hours) {
    const h = Math.floor(hours);
    const m = Math.round((hours - h) * 60);
    return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`;
}

function parseTime(timeString) {
    const [hours, minutes] = timeString.split(':').map(Number);
    return hours + (minutes / 60);
}

// Функция для добавления нового периода
function addInsulinPeriod() {
    const container = document.getElementById('insulin-coefficients-container');
    const periodDiv = document.createElement('div');
    periodDiv.className = 'period-entry mb-3';
    periodDiv.innerHTML = `
        <div class="row">
            <div class="col-md-3">
                <label class="form-label">Начало периода</label>
                <input type="time" class="form-control period-start" required>
            </div>
            <div class="col-md-3">
                <label class="form-label">Коэффициент</label>
                <input type="number" step="0.01" min="0" class="form-control period-coefficient" required>
            </div>
            <div class="col-md-3">
                <label class="form-label">Часы</label>
                <input type="number" step="0.5" min="0" max="24" class="form-control period-hours" required>
            </div>
            <div class="col-md-3">
                <label class="form-label">&nbsp;</label>
                <button type="button" class="btn btn-danger w-100 remove-period">Удалить</button>
            </div>
        </div>
    `;
    container.appendChild(periodDiv);
    updateTotalHours();
}

// Функция для добавления нового периода чувствительности
function addSensitivityPeriod() {
    const container = document.getElementById('insulin-sensitivity-container');
    const periodDiv = document.createElement('div');
    periodDiv.className = 'period-entry mb-3';
    periodDiv.innerHTML = `
        <div class="row">
            <div class="col-md-3">
                <label class="form-label">Начало периода</label>
                <input type="time" class="form-control period-start" required>
            </div>
            <div class="col-md-3">
                <label class="form-label">Чувствительность</label>
                <input type="number" step="0.1" min="0" class="form-control period-sensitivity" required>
            </div>
            <div class="col-md-3">
                <label class="form-label">Часы</label>
                <input type="number" step="0.5" min="0" max="24" class="form-control period-hours" required>
            </div>
            <div class="col-md-3">
                <label class="form-label">&nbsp;</label>
                <button type="button" class="btn btn-danger w-100 remove-period">Удалить</button>
            </div>
        </div>
    `;
    container.appendChild(periodDiv);
    updateSensitivityTotalHours();
}

// Функция для добавления нового периода углеводного соотношения
function addCarbRatioPeriod() {
    const container = document.getElementById('carb-ratio-container');
    const periodDiv = document.createElement('div');
    periodDiv.className = 'period-entry mb-3';
    periodDiv.innerHTML = `
        <div class="row">
            <div class="col-md-3">
                <label class="form-label">Начало периода</label>
                <input type="time" class="form-control period-start" required>
            </div>
            <div class="col-md-3">
                <label class="form-label">Соотношение</label>
                <input type="number" step="0.1" min="0" class="form-control period-ratio" required>
            </div>
            <div class="col-md-3">
                <label class="form-label">Часы</label>
                <input type="number" step="0.5" min="0" max="24" class="form-control period-hours" required>
            </div>
            <div class="col-md-3">
                <label class="form-label">&nbsp;</label>
                <button type="button" class="btn btn-danger w-100 remove-period">Удалить</button>
            </div>
        </div>
    `;
    container.appendChild(periodDiv);
    updateCarbRatioTotalHours();
}

// Функция для удаления периода
function removePeriod(event) {
    const periodDiv = event.target.closest('.period-entry');
    if (periodDiv) {
        periodDiv.remove();
        updateTotalHours();
        updateSensitivityTotalHours();
        updateCarbRatioTotalHours();
    }
}

// Функция для проверки и корректировки периодов
function validatePeriods(type = 'insulin') {
    let periods = [];
    let totalHoursDisplay;
    let containerId;
    
    if (type === 'insulin') {
        periods = getInsulinPeriods();
        totalHoursDisplay = document.getElementById('total-hours-display');
        containerId = 'insulin-coefficients-container';
    } else if (type === 'sensitivity') {
        periods = getSensitivityPeriods();
        totalHoursDisplay = document.getElementById('sensitivity-total-hours-display');
        containerId = 'insulin-sensitivity-container';
    } else if (type === 'carb-ratio') {
        periods = getCarbRatioPeriods();
        totalHoursDisplay = document.getElementById('carb-ratio-total-hours-display');
        containerId = 'carb-ratio-container';
    }
    
    let totalHours = 0;
    
    periods.forEach(period => {
        totalHours += (period.endHour - period.startHour);
    });
    
    if (totalHours === 24) {
        totalHoursDisplay.className = 'text-success';
        totalHoursDisplay.textContent = 'Сумма периодов: 24 часа (правильно)';
        return true;
    } else {
        totalHoursDisplay.className = 'text-danger';
        totalHoursDisplay.textContent = `Текущая сумма: ${totalHours.toFixed(2)} часов из 24 необходимых`;
        showAlert(`Периоды ${type} не покрывают 24 часа. Текущая сумма: ${totalHours.toFixed(2)} часов`, 'danger');
        return false;
    }
}

function getInsulinPeriods() {
    const periodRows = document.querySelectorAll('#insulin-coefficients-container .period-row');
    const periods = [];
    
    periodRows.forEach(row => {
        const startHour = parseInt(row.querySelector('.period-start').value, 10);
        const endHour = parseInt(row.querySelector('.period-end').value, 10);
        const name = row.querySelector('.period-name').value;
        const coefficient = parseFloat(row.querySelector('.period-coefficient').value);
        
        // Validate coefficient value to ensure it's a valid number
        if (isNaN(coefficient)) {
            console.error('Invalid coefficient value found in insulin period');
            return;
        }
        
        periods.push({
            startHour,
            endHour,
            name,
            coefficient: Number(coefficient) // Don't round, preserve full precision
        });
    });
    
    // Sort periods by start hour
    periods.sort((a, b) => a.startHour - b.startHour);
    
    console.log('Retrieved insulin periods:', JSON.stringify(periods)); // Debug log with full JSON
    
    return periods;
}

function getSensitivityPeriods() {
    const periodRows = document.querySelectorAll('#insulin-sensitivity-container .period-row');
    const periods = [];
    
    periodRows.forEach(row => {
        const nameInput = row.querySelector('.sensitivity-name');
        const startInput = row.querySelector('.sensitivity-start');
        const endInput = row.querySelector('.sensitivity-end');
        const coefficientInput = row.querySelector('.sensitivity-coefficient');
        
        if (nameInput && startInput && endInput && coefficientInput) {
            const name = nameInput.value || '';
            const startHour = parseInt(startInput.value, 10) || 0;
            const endHour = parseInt(endInput.value, 10) || 0;
            const coefficient = parseFloat(coefficientInput.value) || 0;
            
            periods.push({
                name: name,
                startHour: startHour,
                endHour: endHour,
                coefficient: coefficient
            });
        }
    });
    
    // Sort periods by start hour
    periods.sort((a, b) => a.startHour - b.startHour);
    
    console.debug("Sensitivity periods:", periods);
    return periods;
}

function getCarbRatioPeriods() {
    const periodRows = document.querySelectorAll('#carb-ratio-container .period-row');
    const periods = [];
    
    periodRows.forEach(row => {
        const nameInput = row.querySelector('.carb-ratio-name');
        const startInput = row.querySelector('.carb-ratio-start');
        const endInput = row.querySelector('.carb-ratio-end');
        const coefficientInput = row.querySelector('.carb-ratio-coefficient');
        
        if (nameInput && startInput && endInput && coefficientInput) {
            const name = nameInput.value || '';
            const startHour = parseInt(startInput.value, 10) || 0;
            const endHour = parseInt(endInput.value, 10) || 0;
            const coefficient = parseFloat(coefficientInput.value) || 0;
            
            periods.push({
                name: name,
                startHour: startHour,
                endHour: endHour,
                coefficient: coefficient
            });
        }
    });
    
    // Sort periods by start hour
    periods.sort((a, b) => a.startHour - b.startHour);
    
    console.debug("Carb ratio periods:", periods);
    return periods;
}

// Отображение определенной страницы/вкладки
function showPage(pageId) {
    // Скрыть все страницы
    document.querySelectorAll('.page').forEach(page => {
        page.classList.remove('active');
    });
    
    // Скрыть все ссылки навигации
    document.querySelectorAll('.nav-link').forEach(link => {
        link.classList.remove('active');
    });
    
    // Показать выбранную страницу
    const selectedPage = document.getElementById(pageId);
    if (selectedPage) {
        selectedPage.classList.add('active');
    }
    
    // Активировать ссылку навигации
    const navLink = document.querySelector(`.nav-link[href="#${pageId}"]`);
    if (navLink) {
        navLink.classList.add('active');
    }
}

function loadUserSettings() {
    const bloodsugarLoadingMsg = document.getElementById('bloodsugar-loading');
    if (bloodsugarLoadingMsg) {
        bloodsugarLoadingMsg.textContent = "Загрузка настроек...";
        bloodsugarLoadingMsg.style.display = 'block';
    }

    return fetch(`${API_BASE_URL}/settings/${currentUserId}`)
        .then(response => {
        if (!response.ok) {
                // If user not found (404), we'll create default settings
                if (response.status === 404) {
                    console.log("User not found, using default settings");
                    return { settings: createDefaultSettings() };
                }
                throw new Error(`Ошибка HTTP: ${response.status}`);
            }
            return response.json();
        })
        .then(responseData => {
            console.log("Loaded user settings response:", JSON.stringify(responseData, null, 2));
            
            // Extract settings from response - could be directly in the response or in a data/settings property
            const data = responseData.data || responseData;
            const settings = data.settings || data;
            
            // If settings is empty or null, create default settings
            if (!settings || Object.keys(settings).length === 0) {
                console.log("Empty settings received, using defaults");
                return createDefaultSettings();
            }
            
            // Создаем глубокую копию настроек
            const settingsCopy = JSON.parse(JSON.stringify(settings));
            
            // Заполняем форму настройками
            populateSettingsForm(settingsCopy);
            
            if (bloodsugarLoadingMsg) {
                bloodsugarLoadingMsg.style.display = 'none';
            }
            
            return settingsCopy;
        })
        .catch(error => {
            console.error('Ошибка загрузки настроек пользователя:', error);
            if (bloodsugarLoadingMsg) {
                bloodsugarLoadingMsg.textContent = `Ошибка загрузки настроек: ${error.message}`;
            }
            // Return default settings on error
            return createDefaultSettings();
        });
}

// Create default settings
function createDefaultSettings() {
    return {
        targetMin: 4.0,
        targetMax: 8.0,
        iobDuration: 4.0,
        insulinPeriods: [{
            startTime: '00:00',
            coefficient: 1.0,
            hours: 24
        }],
        sensitivityPeriods: [{
            startTime: '00:00',
            sensitivity: 2.0,
            hours: 24
        }],
        carbRatioPeriods: [{
            startTime: '00:00',
            ratio: 1.0,
            hours: 24
        }]
    };
}

function populateSettingsForm(settings) {
    if (!settings) {
        settings = createDefaultSettings();
    }
    
    // Basic settings
    document.getElementById('target-min').value = settings.targetMin || 4.0;
    document.getElementById('target-max').value = settings.targetMax || 8.0;
    document.getElementById('iob-duration').value = settings.iobDuration || 4.0;
    
    // Clear existing periods
    document.getElementById('insulin-coefficients-container').innerHTML = '';
    document.getElementById('insulin-sensitivity-container').innerHTML = '';
    document.getElementById('carb-ratio-container').innerHTML = '';
    
    // Add insulin periods
    if (settings.insulinPeriods && settings.insulinPeriods.length > 0) {
        settings.insulinPeriods.forEach(period => {
            addInsulinPeriod();
            const lastPeriod = document.querySelector('#insulin-coefficients-container .period-entry:last-child');
            lastPeriod.querySelector('.period-start').value = period.startTime;
            lastPeriod.querySelector('.period-coefficient').value = period.coefficient;
            lastPeriod.querySelector('.period-hours').value = period.hours;
        });
    } else {
        addInsulinPeriod(); // Add default period
    }
    
    // Add sensitivity periods
    if (settings.sensitivityPeriods && settings.sensitivityPeriods.length > 0) {
        settings.sensitivityPeriods.forEach(period => {
            addSensitivityPeriod();
            const lastPeriod = document.querySelector('#insulin-sensitivity-container .period-entry:last-child');
            lastPeriod.querySelector('.period-start').value = period.startTime;
            lastPeriod.querySelector('.period-sensitivity').value = period.sensitivity;
            lastPeriod.querySelector('.period-hours').value = period.hours;
        });
    } else {
        addSensitivityPeriod(); // Add default period
    }
    
    // Add carb ratio periods
    if (settings.carbRatioPeriods && settings.carbRatioPeriods.length > 0) {
        settings.carbRatioPeriods.forEach(period => {
            addCarbRatioPeriod();
            const lastPeriod = document.querySelector('#carb-ratio-container .period-entry:last-child');
            lastPeriod.querySelector('.period-start').value = period.startTime;
            lastPeriod.querySelector('.period-ratio').value = period.ratio;
            lastPeriod.querySelector('.period-hours').value = period.hours;
        });
    } else {
        addCarbRatioPeriod(); // Add default period
    }
    
    updateTotalHours();
    updateSensitivityTotalHours();
    updateCarbRatioTotalHours();
}

function saveSettings(event) {
    console.log('Save settings called');
    event.preventDefault();
    
    if (!validatePeriods()) {
        console.log('Period validation failed');
        return;
    }
    
    const settings = {
        targetMin: parseFloat(document.getElementById('target-min').value),
        targetMax: parseFloat(document.getElementById('target-max').value),
        iobDuration: parseFloat(document.getElementById('iob-duration').value),
        insulinPeriods: collectPeriods('insulin-coefficients-container', 'coefficient'),
        sensitivityPeriods: collectPeriods('insulin-sensitivity-container', 'sensitivity'),
        carbRatioPeriods: collectPeriods('carb-ratio-container', 'ratio')
    };
    
    console.log('Settings to save:', settings);
    
    fetch(`${API_BASE_URL}/settings/${currentUserId}`, {
            method: 'POST',
            headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(settings)
    })
    .then(response => {
        console.log('Server response:', response);
        if (!response.ok) {
            throw new Error('Ошибка сохранения настроек');
        }
        return response.json();
    })
    .then(data => {
        console.log('Settings saved successfully:', data);
        alert('Настройки успешно сохранены');
        loadUserSettings();
    })
    .catch(error => {
        console.error('Error saving settings:', error);
        alert('Ошибка при сохранении настроек: ' + error.message);
    });
}

function collectPeriods(containerId, valueField) {
    const container = document.getElementById(containerId);
    if (!container) return [];
    
    const periods = [];
    const periodDivs = container.getElementsByClassName('period-row');
    
    for (const div of periodDivs) {
        const startTime = div.querySelector(`.${valueField}-start`).value;
        const endTime = div.querySelector(`.${valueField}-end`).value;
        const name = div.querySelector(`.${valueField}-name`).value;
        const value = parseFloat(div.querySelector(`.${valueField}-coefficient`).value);
        
        if (startTime && endTime && !isNaN(value)) {
            periods.push({
                startHour: parseTime(startTime),
                endHour: parseTime(endTime),
                name,
                coefficient: value
            });
        }
    }
    
    return periods;
}

function updateTotalHours() {
    const total = calculateTotalHours('insulin-coefficients-container');
    const display = document.getElementById('total-hours-display');
    display.textContent = `Текущая сумма: ${total.toFixed(2)} часов из 24 необходимых`;
    display.className = total === 24 ? 'text-success' : 'text-danger';
}

function updateSensitivityTotalHours() {
    const total = calculateTotalHours('insulin-sensitivity-container');
    const display = document.getElementById('sensitivity-total-hours-display');
    display.textContent = `Текущая сумма: ${total.toFixed(2)} часов из 24 необходимых`;
    display.className = total === 24 ? 'text-success' : 'text-danger';
}

function updateCarbRatioTotalHours() {
    const total = calculateTotalHours('carb-ratio-container');
    const display = document.getElementById('carb-ratio-total-hours-display');
    display.textContent = `Текущая сумма: ${total.toFixed(2)} часов из 24 необходимых`;
    display.className = total === 24 ? 'text-success' : 'text-danger';
}

function calculateTotalHours(containerId) {
    const container = document.getElementById(containerId);
    const hourInputs = container.getElementsByClassName('period-hours');
    let total = 0;
    
    for (const input of hourInputs) {
        const value = parseFloat(input.value);
        if (!isNaN(value)) {
            total += value;
        }
    }
    
    return total;
}

function saveBloodSugar(inputId) {
    const value = document.getElementById(inputId).value;
    
    if (!value || isNaN(parseFloat(value))) {
        showAlert('Пожалуйста, введите корректное значение сахара крови', 'danger');
        return;
    }
    
    // Формируем данные для запроса
    const data = {
        userId: currentUserId,
        value: parseFloat(value)
    };
    
    // Отправляем запрос на сервер
    fetch(`${API_BASE_URL}/bloodsugar`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Ошибка сети или сервера');
        }
        return response.json();
    })
    .then(data => {
        // Очищаем поле ввода
        document.getElementById(inputId).value = '';
        
        // Показываем сообщение об успехе
        showAlert('Значение сахара крови успешно сохранено', 'success');
        
        // Обновляем список и график
        loadBloodSugarReadings();
    })
    .catch(error => {
        console.error('Ошибка при сохранении данных:', error);
        showAlert('Ошибка при сохранении данных: ' + error.message, 'danger');
    });
}

/**
 * Attempts to restart the server connection
 */
function restartServerConnection() {
    showAlert('Пытаемся восстановить соединение с сервером...', 'info');
    
    fetch(`${API_BASE_URL}/health`)
        .then(response => {
        if (!response.ok) {
                throw new Error('Сервер не отвечает');
            }
            return response.json();
        })
        .then(data => {
            showAlert('Соединение с сервером восстановлено!', 'success');
            // Reload the current page data
            if (document.getElementById('bloodsugar-list')) {
                loadBloodSugarReadings();
            } else if (document.getElementById('settings-form')) {
                loadUserSettings();
            }
        })
        .catch(error => {
            showAlert('Не удалось восстановить соединение с сервером. Пожалуйста, перезагрузите страницу.', 'danger');
        });
}

function loadBloodSugarReadings() {
    const bloodsugarContainer = document.getElementById('bloodsugar-container');
    if (!bloodsugarContainer) return;
    
    // Show loading message
    bloodsugarContainer.innerHTML = '<div class="text-center"><div class="spinner-border" role="status"></div><p class="mt-2">Загрузка показаний...</p></div>';
    
    // Fetch blood sugar readings
    fetch(`${API_BASE_URL}/bloodsugar/${currentUserId}`)
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            return response.json();
        })
        .then(data => {
            console.log('Raw data from server:', data);
            
            // Handle both old format (data.readings) and new format (direct array)
            let readings = Array.isArray(data) ? data : (data && data.readings ? data.readings : []);
            
            // Process readings and ensure all dates are Date objects
            const processedReadings = readings.map(reading => ({
                ...reading,
                timestamp: new Date(reading.timestamp)
            }));
            
            // Store readings globally for reference in other functions
            window.currentReadings = processedReadings;
            
            // Update the display and chart
            if (processedReadings.length === 0) {
                bloodsugarContainer.innerHTML = '<div class="alert alert-info">У вас пока нет сохраненных показаний.</div>';
            } else {
                bloodsugarContainer.innerHTML = '<ul id="bloodsugar-list" class="list-group mb-4"></ul><div id="chart-container" class="mt-4"><canvas id="bloodsugar-chart"></canvas></div>';
                displayBloodSugarReadings(processedReadings);
                updateBloodSugarChart('bloodsugar-chart', processedReadings);
            }
        })
        .catch(error => {
        console.error('Error loading blood sugar readings:', error);
            
            // Check if it's a timeout error
            if (error.message.includes('timeout') || error.name === 'TimeoutError') {
                bloodsugarContainer.innerHTML = `
                    <div class="alert alert-warning">
                        <p>Превышено время ожидания ответа от сервера.</p>
                        <button class="btn btn-primary mt-2" onclick="restartServerConnection()">Повторить попытку</button>
                    </div>`;
            } else {
                bloodsugarContainer.innerHTML = `<div class="alert alert-danger">Ошибка при загрузке показаний: ${error.message}</div>`;
            }
        });
}

function displayBloodSugarReadings(readings) {
    const bloodSugarList = document.getElementById('bloodsugar-list');
    if (!bloodSugarList) {
        console.error('Element with ID "bloodsugar-list" not found');
        return;
    }
    
    bloodSugarList.innerHTML = '';

    if (!readings || readings.length === 0) {
        bloodSugarList.innerHTML = '<div class="alert alert-info">Нет записей о сахаре крови</div>';
        return;
    }

    // Сортируем показатели по времени (от новых к старым)
    readings.sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp));
    
    readings.forEach(reading => {
        const date = new Date(reading.timestamp);
        const formattedDate = date.toLocaleString('ru-RU', {
            day: '2-digit',
            month: '2-digit',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });

        const listItem = document.createElement('li');
        listItem.className = 'list-group-item d-flex justify-content-between align-items-center';
        
        let valueColorClass = '';
        if (reading.value < 4.0) {
            valueColorClass = 'text-danger';
        } else if (reading.value > 7.8) {
            valueColorClass = 'text-warning';
        } else {
            valueColorClass = 'text-success';
        }

        listItem.innerHTML = `
            <div>
                <span class="timestamp">${formattedDate}</span>
                <span class="source badge bg-secondary">${reading.source || 'Ручной ввод'}</span>
            </div>
            <span class="value ${valueColorClass}">${reading.value.toFixed(1)} ммоль/л</span>
        `;
        
        // Добавляем кнопку удаления
        const deleteButton = document.createElement('button');
        deleteButton.className = 'btn btn-sm btn-danger ms-2';
        deleteButton.innerHTML = '<i class="bi bi-trash"></i> Удалить';
        
        // Use the timestamp as the ID for the reading
        const readingId = date.getTime();
        deleteButton.addEventListener('click', () => deleteBloodSugarReading(readingId));
        
        listItem.appendChild(deleteButton);
        bloodSugarList.appendChild(listItem);
    });
}

function updateBloodSugarChart(chartContainerId, readings) {
    const chartContainer = document.getElementById(chartContainerId);
    if (!chartContainer) {
        console.error('Элемент графика не найден');
        return;
    }
    
    // Если нет данных, скрываем график
    if (!readings || readings.length === 0) {
        chartContainer.innerHTML = '<div class="alert alert-info">Нет данных для отображения графика</div>';
        return;
    }
    
    // Подготавливаем данные для графика
    const chartData = readings.map(reading => ({
        x: new Date(reading.timestamp).getTime(), // Convert to timestamp for better compatibility
        y: reading.value
    }));
    
    // Сортируем по времени
    chartData.sort((a, b) => a.x - b.x);
    
    // Если график уже создан, обновляем данные
    if (bloodSugarChart) {
        bloodSugarChart.data.datasets[0].data = chartData;
        bloodSugarChart.update();
        return;
    }
    
    // Иначе создаем новый график
    const ctx = chartContainer.getContext('2d');
    
    // Register annotation plugin if available
    if (typeof ChartAnnotation !== 'undefined') {
        Chart.register(ChartAnnotation);
    }
    
    try {
    bloodSugarChart = new Chart(ctx, {
        type: 'line',
        data: {
            datasets: [{
                    label: 'Сахар крови (ммоль/л)',
                    data: chartData,
                borderColor: 'rgb(75, 192, 192)',
                    backgroundColor: 'rgba(75, 192, 192, 0.2)',
                    tension: 0.2,
                    fill: true,
                    pointRadius: 5,
                    pointHoverRadius: 7
            }]
        },
        options: {
            responsive: true,
                maintainAspectRatio: false,
                scales: {
                    x: {
                        type: 'time',
                        time: {
                            unit: 'hour',
                            displayFormats: {
                                hour: 'dd.MM HH:mm'
                            }
                        },
                        title: {
                            display: true,
                            text: 'Время'
                        }
                    },
                    y: {
                        min: 2.0,
                        max: 15.0,
                        title: {
                            display: true,
                            text: 'Сахар крови (ммоль/л)'
                        }
                    }
                },
                plugins: {
                    tooltip: {
                        callbacks: {
                            title: function(context) {
                                const date = new Date(context[0].parsed.x);
                                return date.toLocaleString('ru-RU', {
                                    day: '2-digit',
                                    month: '2-digit',
                                    year: 'numeric',
                                    hour: '2-digit',
                                    minute: '2-digit'
                                });
                            }
                        }
                    },
                    annotation: {
                        annotations: {
                            targetRangeMin: {
                                type: 'line',
                                yMin: 4.0,
                                yMax: 4.0,
                                borderColor: 'rgba(255, 99, 132, 0.5)',
                                borderWidth: 2,
                                borderDash: [5, 5],
                                label: {
                                    display: true,
                                    content: 'Мин. норма (4.0)',
                                    position: 'start'
                                }
                            },
                            targetRangeMax: {
                                type: 'line',
                                yMin: 7.8,
                                yMax: 7.8,
                                borderColor: 'rgba(255, 99, 132, 0.5)',
                                borderWidth: 2,
                                borderDash: [5, 5],
                                label: {
                                    display: true,
                                    content: 'Макс. норма (7.8)',
                                    position: 'end'
                                }
                            }
                        }
                    }
                }
            }
        });
    } catch (error) {
        console.error('Error creating chart:', error);
        chartContainer.innerHTML = `<div class="alert alert-danger">Ошибка при создании графика: ${error.message}</div>`;
    }
}

function analyzeFood(e) {
    e.preventDefault();
    
    const form = document.getElementById('food-analysis-form');
    const foodPhotoInput = document.getElementById('food-photo');
    const foodWeightInput = document.getElementById('food-weight');
    
    // Check if photo is provided
    if (!foodPhotoInput.files || foodPhotoInput.files.length === 0) {
        showAlert('Пожалуйста, загрузите фото блюда', 'warning');
        return;
    }
    
    const foodPhoto = foodPhotoInput.files[0];
    const foodWeight = foodWeightInput.value ? parseFloat(foodWeightInput.value) : 0;
    
    // Check file size
    if (foodPhoto.size > 10 * 1024 * 1024) { // 10MB
        showAlert('Размер файла не должен превышать 10MB', 'warning');
        return;
    }
    
    // Show loading indicator
    const analysisResult = document.getElementById('food-analysis-result');
    const analysisContent = document.getElementById('analysis-content');
    analysisContent.innerHTML = '<div class="text-center"><div class="spinner-border" role="status"><span class="visually-hidden">Загрузка...</span></div><p class="mt-3">Анализируем пищу...</p></div>';
    analysisResult.style.display = 'block';
    
    // Create FormData
    const formData = new FormData();
    formData.append('userId', currentUserId);
    formData.append('foodPhoto', foodPhoto);
    if (foodWeight > 0) {
        formData.append('foodWeight', foodWeight);
    }
    
    // Send request to server
    fetch(`${API_BASE_URL}/analyze-food`, {
            method: 'POST',
        body: formData
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        return response.json();
    })
    .then(data => {
        // Handle successful response
        handleAnalysisResponse(data, foodPhoto, "", foodWeight);
    })
    .catch(error => {
        // Handle error
        handleAnalysisError(error);
    });
}

function handleAnalysisResponse(response, foodPhoto, foodInput, foodWeight) {
    const analysisContent = document.getElementById('analysis-content');
    let html = '';
    
    if (response.success) {
        const analysis = response.analysis;
        
        // Create object URL for the photo if provided
        let photoUrl = '';
        if (foodPhoto) {
            photoUrl = URL.createObjectURL(foodPhoto);
        }
        
        html += '<div class="row">';
        
        // Add photo if available
        if (photoUrl) {
            html += `<div class="col-md-4 mb-3">
                        <img src="${photoUrl}" class="img-fluid rounded" alt="Food photo">
                    </div>`;
        }
        
        // Analysis details
        html += `<div class="${photoUrl ? 'col-md-8' : 'col-12'}">
                    <h4>${analysis.dish}</h4>
                    <div class="d-flex justify-content-between mb-2">
                        <span>Углеводы:</span>
                        <strong>${analysis.carbs.toFixed(1)} г</strong>
                    </div>
                    <div class="d-flex justify-content-between mb-2">
                        <span>Доза инсулина:</span>
                        <strong>${analysis.totalInsulin.toFixed(1)} ед.</strong>
                    </div>
                    <div class="d-flex justify-content-between mb-2">
                        <span>Коэффициент времени:</span>
                        <strong>×${analysis.periodCoefficient.toFixed(1)}</strong>
                    </div>
                    <div class="d-flex justify-content-between mb-2">
                        <span>Уверенность:</span>
                        <strong>${analysis.confidence}</strong>
                    </div>`;
        
        // Add weight if provided
        if (foodWeight > 0) {
            html += `<div class="d-flex justify-content-between mb-2">
                        <span>Вес:</span>
                        <strong>${foodWeight} г</strong>
                    </div>`;
        }
        
        html += `</div></div>`;
        
        // Add reasoning
        if (analysis.reasoning) {
            html += `<div class="mt-3 pt-3 border-top">
                        <p><strong>Подробности:</strong></p>
                        <p>${analysis.reasoning}</p>
                    </div>`;
        }
        
        // Dose breakdown
        html += `<div class="mt-3 pt-3 border-top">
                    <h5>Расчет дозы инсулина:</h5>
                    <div class="d-flex justify-content-between mb-2">
                        <span>Инсулин на еду:</span>
                        <span>${analysis.mealInsulin.toFixed(1)} ед.</span>
                    </div>`;
        
        if (analysis.correctionInsulin > 0) {
            html += `<div class="d-flex justify-content-between mb-2">
                        <span>Коррекционный инсулин:</span>
                        <span>${analysis.correctionInsulin.toFixed(1)} ед.</span>
                    </div>`;
        }
        
        html += `<div class="d-flex justify-content-between fw-bold">
                    <span>Итоговая доза:</span>
                    <span>${analysis.totalInsulin.toFixed(1)} ед.</span>
                </div>
            </div>`;
        
    } else {
        html = `<div class="alert alert-danger">Ошибка анализа: ${response.error || 'Неизвестная ошибка'}</div>`;
    }
    
    analysisContent.innerHTML = html;
}

function handleAnalysisError(error) {
    console.error('Ошибка при анализе пищи:', error);
    const analysisContent = document.getElementById('analysis-content');
    analysisContent.innerHTML = `
        <div class="alert alert-danger">
            <p><strong>Ошибка при анализе пищи:</strong></p>
            <p>${error.message || 'Неизвестная ошибка'}</p>
        </div>
    `;
}

// Функция для синхронизации с LibreView
function syncWithLibre() {
    showAlert('Функция синхронизации находится в разработке', 'info');
}

// Функция для отображения сообщений
function showAlert(message, type = 'info') {
    const alertsContainer = document.getElementById('alerts');
    if (!alertsContainer) return;
    
    // Создаем элемент сообщения
    const alertElement = document.createElement('div');
    alertElement.className = `alert alert-${type} alert-dismissible fade show`;
    alertElement.role = 'alert';
    
    alertElement.innerHTML = `
        ${message}
        <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>
    `;
    
    // Очищаем предыдущие сообщения
    alertsContainer.innerHTML = '';
    
    // Добавляем новое сообщение
    alertsContainer.appendChild(alertElement);
    
    // Автоматически скрываем информационные сообщения через 5 секунд
    if (type === 'info' || type === 'success') {
    setTimeout(() => {
            const alert = alertElement;
            if (alert && alert.parentNode) {
                const bsAlert = new bootstrap.Alert(alert);
                bsAlert.close();
            }
    }, 5000);
    }
}

/**
 * Deletes a blood sugar reading by ID
 * @param {string} readingId - The ID of the reading to delete
 */
function deleteBloodSugarReading(readingId) {
    if (!readingId) {
        console.error('No reading ID provided for deletion');
        return;
    }
    
    // Find the reading with the given ID from our data
    const bloodsugarList = document.getElementById('bloodsugar-list');
    if (!bloodsugarList) {
        console.error('Blood sugar list element not found');
        return;
    }
    
    // Ask for confirmation before deleting
    if (!confirm('Вы уверены, что хотите удалить это показание?')) {
        return;
    }

    // In our application, the timestamp is being used as the ID
    // Find the corresponding reading from the data we have
    const readings = window.currentReadings || [];
    const readingToDelete = readings.find(r => new Date(r.timestamp).getTime() === readingId);
    
    if (!readingToDelete) {
        showAlert('Ошибка: Не удалось найти показание для удаления', 'danger');
        return;
    }
    
    console.log('Reading to delete:', readingToDelete);
    
    // Send request to the server
    fetch(`${API_BASE_URL}/bloodsugar`, {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            userId: currentUserId,
            timestamp: readingToDelete.timestamp.toISOString ? 
                      readingToDelete.timestamp.toISOString() : 
                      readingToDelete.timestamp
        })
    })
    .then(response => {
        if (!response.ok) {
            return response.json().then(data => {
                throw new Error(data.error || `HTTP error! Status: ${response.status}`);
            });
        }
        return response.json();
    })
    .then(data => {
        showAlert('Показание успешно удалено', 'success');
        loadBloodSugarReadings(); // Reload the list after deletion
    })
    .catch(error => {
        console.error('Ошибка при удалении показания:', error);
        showAlert('Ошибка при удалении показания: ' + error.message, 'danger');
    });
} 