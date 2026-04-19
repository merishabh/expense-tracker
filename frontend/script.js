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
    if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || `Request failed: ${response.status}`);
    }
    return response.json();
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
                    <strong>${item.label}</strong>
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
            <td>${tx.date_time ? new Date(tx.date_time).toLocaleString() : '-'}</td>
            <td>${tx.vendor || '-'}</td>
            <td>${tx.category || 'Other'}</td>
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
                ? `${topCategory.label} at ${formatCurrency(topCategory.amount)}`
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
    loadDashboard();
});
