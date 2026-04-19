function formatCategory(name) {
    return (name || '').replace(/_/g, ' ');
}

function formatCurrency(value) {
    return new Intl.NumberFormat('en-IN', {
        style: 'currency',
        currency: 'INR',
        maximumFractionDigits: 2
    }).format(value || 0);
}

const dashboardState = {
    selectedCategory: null,
    categories: [],
    periodTransactions: [],
    lastTenDaysTransactions: []
};

async function fetchJSON(path) {
    const response = await fetch(path);
    return handleJSONResponse(response);
}

async function sendJSON(path, options) {
    const response = await fetch(path, options);
    return handleJSONResponse(response);
}

async function handleJSONResponse(response) {
    if (response.status === 401) {
        window.location.href = '/auth/signin';
        return;
    }
    if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || `Request failed: ${response.status}`);
    }
    return response.json();
}

function formatDateTime(value) {
    if (!value) return '-';
    return new Date(value).toLocaleString();
}

function getUploadErrorMessage(error) {
    const message = String(error?.message || '');
    if (message.includes('404')) {
        return 'Upload API not found. Restart the Go server so the new /api/import/google-pay route is loaded.';
    }
    if (message.includes('401')) {
        return 'Your session expired. Sign in again and retry the upload.';
    }
    return `Import failed: ${message}`;
}

function renderBreakdown(containerId, items, emptyMessage) {
    const container = document.getElementById(containerId);
    if (!container) return;

    if (!items.length) {
        container.innerHTML = `<p class="empty">${emptyMessage}</p>`;
        return;
    }

    container.innerHTML = items.map(item => `
        <div class="stack-row">
            <div>
                <strong>${item.label}</strong>
                <span>${item.count} txns</span>
            </div>
            <strong>${formatCurrency(item.amount)}</strong>
        </div>
    `).join('');
}

function renderCategoryBreakdown(items) {
    const container = document.getElementById('categoryList');
    if (!container) return;

    if (!items.length) {
        container.innerHTML = '<p class="empty">No categories found.</p>';
        return;
    }

    container.innerHTML = items.map(item => {
        const isActive = dashboardState.selectedCategory && dashboardState.selectedCategory.toLowerCase() === String(item.label || '').toLowerCase();
        return `
            <button class="stack-row clickable ${isActive ? 'active' : ''}" data-category="${item.label}" type="button">
                <div>
                    <strong>${formatCategory(item.label)}</strong>
                    <span>${item.count} txns</span>
                </div>
                <strong>${formatCurrency(item.amount)}</strong>
            </button>
        `;
    }).join('');
}

function renderTrend(items) {
    const container = document.getElementById('trendList');
    if (!container) return;

    if (!items.length) {
        container.innerHTML = '<p class="empty">No trend data for this period.</p>';
        return;
    }

    const maxAmount = Math.max(...items.map(item => item.amount), 1);
    container.innerHTML = items.map(item => {
        const width = `${Math.max((item.amount / maxAmount) * 100, 8)}%`;
        return `
            <div class="trend-row">
                <div class="trend-meta">
                    <strong>${item.date}</strong>
                    <span>${item.count} txns</span>
                </div>
                <div class="trend-bar-wrap">
                    <div class="trend-bar" style="width:${width}"></div>
                </div>
                <strong class="trend-value">${formatCurrency(item.amount)}</strong>
            </div>
        `;
    }).join('');
}

function setTransactionsTitle(title) {
    const titleEl = document.getElementById('transactionsTitle');
    if (titleEl) {
        titleEl.textContent = title;
    }
}

function scrollToTransactions() {
    const table = document.getElementById('transactionTable');
    if (!table) return;
    const section = table.closest('.card') || table;
    section.scrollIntoView({ behavior: 'smooth', block: 'start' });
}

function getTxCategory(tx) {
    const raw = String(tx?.category || '').trim();
    return raw || 'Other';
}

function getFilteredTransactionsByCategory(category) {
    const selected = String(category || '').trim().toLowerCase();
    return dashboardState.periodTransactions.filter(tx => getTxCategory(tx).toLowerCase() === selected);
}

function renderTransactions(transactions, emptyMessage = 'No transactions found.') {
    const container = document.getElementById('transactionTable');
    if (!container) return;

    if (!transactions.length) {
        container.innerHTML = `<p class="empty">${emptyMessage}</p>`;
        return;
    }

    const rows = transactions.map(tx => `
        <tr>
            <td>${formatDateTime(tx.date_time)}</td>
            <td>${tx.vendor || '-'}</td>
            <td>${formatCategory(tx.category || 'Other')}</td>
            <td>${tx.type || '-'}</td>
            <td>${formatCurrency(tx.amount)}</td>
        </tr>
    `).join('');

    container.innerHTML = `
        <table>
            <thead>
                <tr>
                    <th>Date</th>
                    <th>Merchant</th>
                    <th>Category</th>
                    <th>Source</th>
                    <th>Amount</th>
                </tr>
            </thead>
            <tbody>${rows}</tbody>
        </table>
    `;
}

function renderMonthlyComparison(comparison) {
    const container = document.getElementById('monthComparison');
    if (!container) return;

    const isUp = (comparison.delta_amount || 0) >= 0;
    const deltaPrefix = isUp ? '+' : '';

    container.innerHTML = `
        <div class="comparison-grid">
            <div class="comparison-card">
                <span>Current Month</span>
                <strong>${formatCurrency(comparison.current_month_amount)}</strong>
                <small>${comparison.current_month_count || 0} txns</small>
            </div>
            <div class="comparison-card">
                <span>Last Month</span>
                <strong>${formatCurrency(comparison.last_month_amount)}</strong>
                <small>${comparison.last_month_count || 0} txns</small>
            </div>
        </div>
        <div class="delta-banner ${isUp ? 'up' : 'down'}">
            <strong>${deltaPrefix}${formatCurrency(comparison.delta_amount)}</strong>
            <span>${deltaPrefix}${(comparison.delta_percent || 0).toFixed(1)}% vs last month</span>
        </div>
    `;
}

function renderHighlights(summary, comparison, categories, transactions) {
    const container = document.getElementById('analyticsHighlights');
    if (!container) return;

    const topCategory = (categories.items || [])[0];
    const latestTransaction = (transactions.transactions || [])[0];

    const highlights = [
        {
            title: 'Largest Category',
            body: topCategory
                ? `${formatCategory(topCategory.label)} at ${formatCurrency(topCategory.amount)}`
                : 'No category data available'
        },
        {
            title: 'Top Merchant This Month',
            body: comparison.top_merchant_this_month
                ? `${comparison.top_merchant_this_month} at ${formatCurrency(comparison.top_merchant_spend)}`
                : 'No merchant concentration yet'
        },
        {
            title: 'Latest Transaction',
            body: latestTransaction
                ? `${latestTransaction.vendor || 'Unknown'} for ${formatCurrency(latestTransaction.amount)}`
                : 'No recent transactions'
        },
        {
            title: 'Review Queue',
            body: `${summary.uncategorized_count || 0} transactions still need categorization review`
        }
    ];

    container.innerHTML = highlights.map(item => `
        <div class="highlight-card">
            <span>${item.title}</span>
            <strong>${item.body}</strong>
        </div>
    `).join('');
}

async function loadDashboard() {
    const period = document.getElementById('periodSelect')?.value || 'THIS_MONTH';

    try {
        const [summary, categories, trend, transactions, monthlyComparison, lastTenDays] = await Promise.all([
            fetchJSON(`/api/summary/total?period=${period}`),
            fetchJSON(`/api/summary/category?period=${period}`),
            fetchJSON('/api/summary/trend/last-10-days'),
            fetchJSON(`/api/transactions?period=${period}&limit=500`),
            fetchJSON('/api/summary/monthly-comparison'),
            fetchJSON('/api/transactions/last-10-days')
        ]);

        document.getElementById('totalAmount').textContent = formatCurrency(summary.total_amount);
        document.getElementById('transactionCount').textContent = String(summary.transaction_count || 0);
        document.getElementById('averageAmount').textContent = formatCurrency(summary.average_amount);
        document.getElementById('uncategorizedCount').textContent = String(summary.uncategorized_count || 0);

        dashboardState.categories = categories.items || [];
        dashboardState.periodTransactions = transactions.transactions || [];
        dashboardState.lastTenDaysTransactions = lastTenDays.transactions || [];
        if (dashboardState.selectedCategory) {
            const categoryStillExists = dashboardState.categories.some(item =>
                String(item.label || '').toLowerCase() === dashboardState.selectedCategory.toLowerCase()
            );
            if (!categoryStillExists) {
                dashboardState.selectedCategory = null;
            }
        }

        renderCategoryBreakdown(dashboardState.categories);
        populateRangeCategoryDropdown(dashboardState.categories);
        renderMonthlyComparison(monthlyComparison);
        renderTrend(trend.items || []);
        if (dashboardState.selectedCategory) {
            const filtered = getFilteredTransactionsByCategory(dashboardState.selectedCategory);
            setTransactionsTitle(`Transactions: ${dashboardState.selectedCategory}`);
            renderTransactions(filtered);
        } else {
            setTransactionsTitle('Last 10 Days Transactions');
            renderTransactions(dashboardState.lastTenDaysTransactions);
        }
        renderHighlights(summary, monthlyComparison, categories, transactions);
    } catch (error) {
        console.error(error);
        document.body.innerHTML = `<main class="page"><section class="card"><h2>Dashboard failed to load</h2><p>${error.message}</p></section></main>`;
    }
}

function populateRangeCategoryDropdown(categories) {
    const select = document.getElementById('rangeCategory');
    if (!select) return;
    const current = select.value;
    select.innerHTML = '<option value="">All Categories</option>';
    categories.forEach(item => {
        const opt = document.createElement('option');
        opt.value = item.label;
        opt.textContent = formatCategory(item.label);
        select.appendChild(opt);
    });
    if (current) select.value = current;
}

function renderRangeResults(data) {
    const summary = document.getElementById('rangeSummary');
    const table = document.getElementById('rangeTable');
    if (!summary || !table) return;

    const transactions = data.transactions || [];
    const total = transactions.reduce((sum, tx) => sum + (tx.amount || 0), 0);

    summary.style.display = 'flex';
    summary.innerHTML = `
        <div class="range-stat"><span>Transactions</span><strong>${data.count || 0}</strong></div>
        <div class="range-stat"><span>Total Spend</span><strong>${formatCurrency(total)}</strong></div>
        <div class="range-stat"><span>Average</span><strong>${formatCurrency(transactions.length ? total / transactions.length : 0)}</strong></div>
    `;

    if (!transactions.length) {
        table.innerHTML = '<p class="empty">No transactions found for the selected range.</p>';
        return;
    }

    const rows = transactions.map(tx => `
        <tr>
            <td>${formatDateTime(tx.date_time)}</td>
            <td>${tx.vendor || '-'}</td>
            <td>${formatCategory(tx.category || 'Other')}</td>
            <td>${tx.type || '-'}</td>
            <td>${formatCurrency(tx.amount)}</td>
        </tr>
    `).join('');

    table.innerHTML = `
        <table>
            <thead>
                <tr><th>Date</th><th>Merchant</th><th>Category</th><th>Source</th><th>Amount</th></tr>
            </thead>
            <tbody>${rows}</tbody>
        </table>
    `;
}

function renderGooglePayImportResult(summary) {
    const result = document.getElementById('googlePayImportResult');
    if (!result) return;

    result.style.display = 'grid';
    result.innerHTML = `
        <div class="range-stat"><span>Imported</span><strong>${summary.imported_count || 0}</strong></div>
        <div class="range-stat"><span>Skipped Old</span><strong>${summary.skipped_old_count || 0}</strong></div>
        <div class="range-stat"><span>Skipped Status</span><strong>${summary.skipped_status_count || 0}</strong></div>
        <div class="range-stat"><span>Skipped Invalid</span><strong>${summary.skipped_invalid_count || 0}</strong></div>
        <div class="range-stat"><span>Latest Stored</span><strong>${formatDateTime(summary.latest_stored_at)}</strong></div>
        <div class="range-stat"><span>Latest Imported</span><strong>${formatDateTime(summary.latest_imported_at)}</strong></div>
    `;
}

async function uploadGooglePayHistory() {
    const fileInput = document.getElementById('googlePayFile');
    const button = document.getElementById('googlePayUpload');
    const result = document.getElementById('googlePayImportResult');

    if (!fileInput?.files?.length) {
        alert('Please choose the Google Pay HTML file first.');
        return;
    }

    const file = fileInput.files[0];
    const formData = new FormData();
    formData.append('file', file);

    button.disabled = true;
    button.textContent = 'Uploading…';
    if (result) {
        result.style.display = 'block';
        result.innerHTML = '<p class="empty">Import in progress...</p>';
    }

    try {
        const response = await sendJSON('/api/import/google-pay', {
            method: 'POST',
            body: formData
        });

        if (response?.summary) {
            renderGooglePayImportResult(response.summary);
        }

        fileInput.value = '';
        await loadDashboard();
    } catch (error) {
        console.error(error);
        if (result) {
            result.style.display = 'block';
            result.innerHTML = `<p class="empty">${getUploadErrorMessage(error)}</p>`;
        }
    } finally {
        button.disabled = false;
        button.textContent = 'Upload History';
    }
}

async function searchByRange() {
    const from = document.getElementById('rangeFrom')?.value;
    const to = document.getElementById('rangeTo')?.value;
    const category = document.getElementById('rangeCategory')?.value || '';
    const btn = document.getElementById('rangeSearch');

    if (!from || !to) {
        alert('Please select both From and To dates.');
        return;
    }
    if (from > to) {
        alert('"From" date must be before "To" date.');
        return;
    }

    btn.disabled = true;
    btn.textContent = 'Loading…';

    try {
        let url = `/api/transactions/range?from=${from}&to=${to}&limit=500`;
        if (category) url += `&category=${encodeURIComponent(category)}`;

        const data = await fetchJSON(url);
        if (data) renderRangeResults(data);
    } catch (err) {
        console.error(err);
        const table = document.getElementById('rangeTable');
        if (table) table.innerHTML = `<p class="empty">Error: ${err.message}</p>`;
    } finally {
        btn.disabled = false;
        btn.textContent = 'Search';
    }
}

document.addEventListener('DOMContentLoaded', () => {
    const periodSelect = document.getElementById('periodSelect');
    if (periodSelect) {
        periodSelect.addEventListener('change', () => {
            dashboardState.selectedCategory = null;
            loadDashboard();
        });
    }

    const categoryList = document.getElementById('categoryList');
    if (categoryList) {
        categoryList.addEventListener('click', async (event) => {
            const row = event.target.closest('[data-category]');
            if (!row) return;

            const clickedCategory = row.getAttribute('data-category');
            if (!clickedCategory) return;

            try {
                if (dashboardState.selectedCategory && dashboardState.selectedCategory.toLowerCase() === clickedCategory.toLowerCase()) {
                    dashboardState.selectedCategory = null;
                    setTransactionsTitle('Last 10 Days Transactions');
                    renderTransactions(dashboardState.lastTenDaysTransactions);
                    renderCategoryBreakdown(dashboardState.categories);
                    scrollToTransactions();
                    return;
                }

                dashboardState.selectedCategory = clickedCategory;
                const filtered = getFilteredTransactionsByCategory(clickedCategory);
                setTransactionsTitle(`Transactions: ${clickedCategory}`);
                renderTransactions(
                    filtered,
                    `No transactions found for "${clickedCategory}" in the selected period.`
                );
                renderCategoryBreakdown(dashboardState.categories);
                scrollToTransactions();
            } catch (error) {
                console.error(error);
            }
        });
    }
    const rangeSearch = document.getElementById('rangeSearch');
    if (rangeSearch) {
        rangeSearch.addEventListener('click', searchByRange);
    }

    const googlePayUpload = document.getElementById('googlePayUpload');
    if (googlePayUpload) {
        googlePayUpload.addEventListener('click', uploadGooglePayHistory);
    }

    // default date range to current month
    const today = new Date();
    const firstOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
    const fmt = d => d.toISOString().slice(0, 10);
    const fromInput = document.getElementById('rangeFrom');
    const toInput = document.getElementById('rangeTo');
    if (fromInput) fromInput.value = fmt(firstOfMonth);
    if (toInput) toInput.value = fmt(today);

    loadDashboard();
});
