function formatCurrency(value) {
    return new Intl.NumberFormat('en-IN', {
        style: 'currency',
        currency: 'INR',
        maximumFractionDigits: 2
    }).format(value || 0);
}

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

function renderTrend(items) {
    const container = document.getElementById('trendList');
    if (!container) return;

    if (!items.length) {
        container.innerHTML = '<p class="empty">No trend data for this period.</p>';
        return;
    }

    container.innerHTML = items.map(item => `
        <div class="stack-row">
            <div>
                <strong>${item.date}</strong>
                <span>${item.count} txns</span>
            </div>
            <strong>${formatCurrency(item.amount)}</strong>
        </div>
    `).join('');
}

function renderTransactions(transactions) {
    const container = document.getElementById('transactionTable');
    if (!container) return;

    if (!transactions.length) {
        container.innerHTML = '<p class="empty">No transactions found.</p>';
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

async function loadDashboard() {
    const period = document.getElementById('periodSelect')?.value || 'THIS_MONTH';

    try {
        const [summary, categories, sources, trend, transactions] = await Promise.all([
            fetchJSON(`/api/summary/total?period=${period}`),
            fetchJSON(`/api/summary/category?period=${period}`),
            fetchJSON(`/api/summary/source?period=${period}`),
            fetchJSON(`/api/summary/trend?period=${period}`),
            fetchJSON(`/api/transactions?period=${period}&limit=25`)
        ]);

        document.getElementById('totalAmount').textContent = formatCurrency(summary.total_amount);
        document.getElementById('transactionCount').textContent = String(summary.transaction_count || 0);
        document.getElementById('averageAmount').textContent = formatCurrency(summary.average_amount);
        document.getElementById('uncategorizedCount').textContent = String(summary.uncategorized_count || 0);

        renderBreakdown('categoryList', categories.items || [], 'No categories found.');
        renderBreakdown('sourceList', sources.items || [], 'No sources found.');
        renderTrend(trend.items || []);
        renderTransactions(transactions.transactions || []);
    } catch (error) {
        console.error(error);
        document.body.innerHTML = `<main class="page"><section class="card"><h2>Dashboard failed to load</h2><p>${error.message}</p></section></main>`;
    }
}

document.addEventListener('DOMContentLoaded', () => {
    const periodSelect = document.getElementById('periodSelect');
    if (periodSelect) {
        periodSelect.addEventListener('change', loadDashboard);
    }
    loadDashboard();
});
