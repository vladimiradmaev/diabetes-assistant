// Загрузка приложения, когда DOM полностью загружен
document.addEventListener('DOMContentLoaded', function() {
  // Проверка доступности Chart.js
  if (typeof Chart === 'undefined') {
    // Загрузка Chart.js, если он еще не загружен
    const chartScript = document.createElement('script');
    chartScript.src = 'https://cdn.jsdelivr.net/npm/chart.js@3.7.1/dist/chart.min.js';
    chartScript.onload = loadAnnotationPlugin;
    document.head.appendChild(chartScript);
  } else {
    // Chart.js уже доступен, загрузка плагина аннотаций
    loadAnnotationPlugin();
  }
  
  function loadAnnotationPlugin() {
    // Загрузка плагина аннотаций, если он еще не загружен
    if (typeof ChartAnnotation === 'undefined') {
      const annotationScript = document.createElement('script');
      annotationScript.src = 'chartjs-plugin-annotation.min.js';
      annotationScript.onload = initializeApp;
      document.head.appendChild(annotationScript);
    } else {
      // Плагин аннотаций уже доступен
      initializeApp();
    }
  }
  
  function initializeApp() {
    // Регистрация плагина аннотаций в Chart.js
    if (typeof Chart !== 'undefined' && typeof ChartAnnotation !== 'undefined') {
      Chart.register(ChartAnnotation);
    }
    
    // Загрузка основного скрипта приложения
    const appScript = document.createElement('script');
    appScript.src = 'app.js';
    document.head.appendChild(appScript);
  }
});

// Utility functions for the Diabetes Assistant application

// Format a date as a readable string
function formatDate(date) {
    return new Date(date).toLocaleString();
}

// Calculate insulin dose based on blood sugar, target, and sensitivity factor
function calculateInsulinDose(currentBG, targetBG, sensitivityFactor) {
    if (!currentBG || !targetBG || !sensitivityFactor) return 0;
    
    const difference = currentBG - targetBG;
    if (difference <= 0) return 0;
    
    return Math.round(difference / sensitivityFactor * 10) / 10;
}

// Calculate insulin for carbs based on carb ratio
function calculateInsulinForCarbs(carbs, carbRatio) {
    if (!carbs || !carbRatio) return 0;
    return Math.round(carbs / carbRatio * 10) / 10;
}

// Get time period (morning, afternoon, evening, night)
function getTimePeriod() {
    const hour = new Date().getHours();
    
    if (hour >= 6 && hour < 12) return 'morning';
    if (hour >= 12 && hour < 18) return 'afternoon';
    if (hour >= 18 && hour < 24) return 'evening';
    return 'night';
}

// Apply base insulin coefficient based on time period
function applyTimeBasedCoefficient(insulin, baseCoefficients) {
    if (!insulin || !baseCoefficients) return insulin;
    
    const period = getTimePeriod();
    const coefficient = baseCoefficients[period] || 1.0;
    
    return Math.round(insulin * coefficient * 10) / 10;
}

// Create a notification
function createNotification(message, type = 'info') {
    if (!("Notification" in window)) {
        console.log("This browser does not support desktop notification");
        return;
    }
    
    if (Notification.permission === "granted") {
        new Notification("Помощник диабетика", {
            body: message,
            icon: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAAgY0hSTQAAeiYAAICEAAD6AAAAgOgAAHUwAADqYAAAOpgAABdwnLpRPAAAAFdQTFRFAAAA/wAAt7e3mpqaZmZmMzMz/wAAZgAAtlpaZjMzmQAAmWYzZjMA/zMzzJmZmWYAMwAAt5lmMzMAzGZm/2ZmmTMAZgAAzDMzZjNmmZmZ////Ri44cwAAAAF0Uk5TAEDm2GYAAAABYktHRAsf18TAAAAACXBIWXMAAC4jAAAuIwF4pT92AAAAB3RJTUUH5gQUBDgSy6n9OgAAAJRJREFUOMvF01kSwyAIBmCjiSbu0fsftRYnlsxk+vSEBxX5FURdEARBdF07DCPpFkyT996FaA4wxvyMi8zlgFJNzZDxIgDe29qQg1ICaq1PBjGuuQQZYwdD732/IIeZAuScS45zPhj0GRd49gIppZMhBDFMDEOhqBQEP4FWIbSttw6F/2aMTfPe+4Y3I/nL6wU/gC/gLwA+ABu5J0k1h9jE"
        });
    } else if (Notification.permission !== "denied") {
        Notification.requestPermission();
    }
}

// Export functions to make them available globally
window.formatDate = formatDate;
window.calculateInsulinDose = calculateInsulinDose;
window.calculateInsulinForCarbs = calculateInsulinForCarbs;
window.getTimePeriod = getTimePeriod;
window.applyTimeBasedCoefficient = applyTimeBasedCoefficient;
window.createNotification = createNotification; 