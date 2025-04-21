/**
 * Chart.js Luxon Adapter initialization
 * This file must be loaded after Chart.js and Luxon, but before any chart is created
 */

// Define the FORMATS object for Luxon
const FORMATS = {
    datetime: luxon.DateTime.DATETIME_MED_WITH_SECONDS,
    millisecond: 'h:mm:ss.SSS a',
    second: luxon.DateTime.TIME_WITH_SECONDS,
    minute: luxon.DateTime.TIME_SIMPLE,
    hour: { hour: 'numeric' },
    day: { day: 'numeric', month: 'short' },
    week: 'DD',
    month: { month: 'short', year: 'numeric' },
    quarter: "'Q'q - yyyy",
    year: { year: 'numeric' }
};

// Manual implementation of the Luxon adapter for Chart.js
Chart._adapters._date.override({
    _id: 'luxon',
    
    /**
     * Creates a DateTime instance from a timestamp
     */
    _create: function(time) {
        return luxon.DateTime.fromMillis(time, this.options);
    },
    
    formats: function() {
        return FORMATS;
    },
    
    parse: function(value, format) {
        const options = this.options;
        
        if (value === null || typeof value === 'undefined') {
            return null;
        }
        
        const type = typeof value;
        if (type === 'number') {
            value = this._create(value);
        } else if (type === 'string') {
            if (typeof format === 'string') {
                value = luxon.DateTime.fromFormat(value, format, options);
            } else {
                value = luxon.DateTime.fromISO(value, options);
            }
        } else if (value instanceof Date) {
            value = luxon.DateTime.fromJSDate(value, options);
        } else if (type === 'object' && !(value instanceof luxon.DateTime)) {
            value = luxon.DateTime.fromObject(value);
        }
        
        return value.isValid ? value.valueOf() : null;
    },
    
    format: function(time, format) {
        const datetime = this._create(time);
        return typeof format === 'string'
            ? datetime.toFormat(format, this.options)
            : datetime.toLocaleString(format);
    },
    
    add: function(time, amount, unit) {
        const args = {};
        args[unit] = amount;
        return this._create(time).plus(args).valueOf();
    },
    
    diff: function(max, min, unit) {
        return this._create(max).diff(this._create(min)).as(unit).valueOf();
    },
    
    startOf: function(time, unit, weekday) {
        if (unit === 'isoWeek') {
            weekday = Math.trunc(Math.min(Math.max(0, weekday), 6));
            const dateTime = this._create(time);
            return dateTime
                .minus({ days: (dateTime.weekday - weekday + 7) % 7 })
                .startOf('day')
                .valueOf();
        }
        return unit ? this._create(time).startOf(unit).valueOf() : time;
    },
    
    endOf: function(time, unit) {
        return this._create(time).endOf(unit).valueOf();
    }
});

console.log('Chart.js Luxon adapter initialized successfully'); 