class Calendar {
    constructor(containerId) {
        this.container = document.getElementById(containerId);
        this.date = new Date();
        this.currentMonth = this.date.getMonth();
        this.currentYear = this.date.getFullYear();
        
        this.monthNames = ["January", "February", "March", "April", "May", "June",
            "July", "August", "September", "October", "November", "December"
        ];
        
        this.init();
    }

    init() {
        this.render();
        this.addEventListeners();
    }

    render() {
        this.container.innerHTML = '';

        // Header
        const header = document.createElement('div');
        header.className = 'calendar-header';

        const prevBtn = document.createElement('button');
        prevBtn.innerText = '<';
        prevBtn.className = 'calendar-nav-btn';
        prevBtn.onclick = () => this.prevMonth();

        const monthYear = document.createElement('span');
        monthYear.className = 'calendar-month-year';
        monthYear.innerText = `${this.monthNames[this.currentMonth]} ${this.currentYear}`;

        const nextBtn = document.createElement('button');
        nextBtn.innerText = '>';
        nextBtn.className = 'calendar-nav-btn';
        nextBtn.onclick = () => this.nextMonth();

        header.appendChild(prevBtn);
        header.appendChild(monthYear);
        header.appendChild(nextBtn);
        this.container.appendChild(header);

        // Days of week
        const daysHeader = document.createElement('div');
        daysHeader.className = 'calendar-days-header';
        const days = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa'];
        days.forEach(day => {
            const dayEl = document.createElement('div');
            dayEl.innerText = day;
            daysHeader.appendChild(dayEl);
        });
        this.container.appendChild(daysHeader);

        // Calendar Grid
        const grid = document.createElement('div');
        grid.className = 'calendar-grid';

        const firstDay = new Date(this.currentYear, this.currentMonth, 1).getDay();
        const daysInMonth = new Date(this.currentYear, this.currentMonth + 1, 0).getDate();

        // Empty cells for days before start of month
        for (let i = 0; i < firstDay; i++) {
            const emptyCell = document.createElement('div');
            emptyCell.className = 'calendar-day empty';
            grid.appendChild(emptyCell);
        }

        // Days
        const today = new Date();
        for (let i = 1; i <= daysInMonth; i++) {
            const dayCell = document.createElement('div');
            dayCell.className = 'calendar-day';
            dayCell.innerText = i;

            if (i === today.getDate() && 
                this.currentMonth === today.getMonth() && 
                this.currentYear === today.getFullYear()) {
                dayCell.classList.add('today');
            }

            grid.appendChild(dayCell);
        }

        this.container.appendChild(grid);
    }

    prevMonth() {
        this.currentMonth--;
        if (this.currentMonth < 0) {
            this.currentMonth = 11;
            this.currentYear--;
        }
        this.render();
    }

    nextMonth() {
        this.currentMonth++;
        if (this.currentMonth > 11) {
            this.currentMonth = 0;
            this.currentYear++;
        }
        this.render();
    }
}
