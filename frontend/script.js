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
    lastTenDaysTransactions: [],
    txMap: {}
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

function registerTxMap(transactions) {
    transactions.forEach(tx => { if (tx.id) dashboardState.txMap[tx.id] = tx; });
}

function renderTransactions(transactions, emptyMessage = 'No transactions found.') {
    registerTxMap(transactions);
    const container = document.getElementById('transactionTable');
    if (!container) return;

    if (!transactions.length) {
        container.innerHTML = `<p class="empty">${emptyMessage}</p>`;
        return;
    }

    const rows = transactions.map(tx => `
        <tr class="tx-row" data-id="${tx.id || ''}">
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
        document.getElementById('averageAmount').textContent = formatCurrency(summary.gross_expense);
        document.getElementById('uncategorizedCount').textContent = formatCurrency(summary.credit_amount);

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

    registerTxMap(transactions);
    const rows = transactions.map(tx => `
        <tr class="tx-row" data-id="${tx.id || ''}">
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
    return `
        <div class="range-stat"><span>Processed</span><strong>${summary.processed_count || 0}${summary.total_blocks ? ` / ${summary.total_blocks}` : ''}</strong></div>
        <div class="range-stat"><span>Imported</span><strong>${summary.imported_count || 0}</strong></div>
        <div class="range-stat"><span>Skipped Old</span><strong>${summary.skipped_old_count || 0}</strong></div>
        <div class="range-stat"><span>Skipped Status</span><strong>${summary.skipped_status_count || 0}</strong></div>
        <div class="range-stat"><span>Skipped Invalid</span><strong>${summary.skipped_invalid_count || 0}</strong></div>
        <div class="range-stat"><span>Batches</span><strong>${summary.batch_count || 0}</strong></div>
        <div class="range-stat"><span>Latest Stored</span><strong>${formatDateTime(summary.latest_stored_at)}</strong></div>
        <div class="range-stat"><span>Latest Imported</span><strong>${formatDateTime(summary.latest_imported_at)}</strong></div>
    `;
}

function renderGooglePayImportJob(job) {
    const result = document.getElementById('googlePayImportResult');
    if (!result) return;

    const summary = job?.summary || {};
    const status = String(job?.status || 'queued');
    const title = status === 'completed'
        ? 'Import completed.'
        : status === 'failed'
            ? `Import failed: ${job.error || 'Unknown error'}`
            : 'Import running in background...';

    result.style.display = 'block';
    result.innerHTML = `
        <p class="empty">${title}</p>
        <div class="import-result-grid">
            ${renderGooglePayImportResult(summary)}
        </div>
    `;
}

async function waitForGooglePayImport(jobId) {
    while (true) {
        const response = await fetchJSON(`/api/import/google-pay?id=${encodeURIComponent(jobId)}`);
        const job = response?.import;
        if (job) {
            renderGooglePayImportJob(job);
        }

        if (!job || job.status === 'completed') {
            return job;
        }
        if (job.status === 'failed') {
            throw new Error(job.error || 'Google Pay import failed');
        }

        await new Promise(resolve => setTimeout(resolve, 1500));
    }
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
    button.textContent = 'Queueing…';
    if (result) {
        result.style.display = 'block';
        result.innerHTML = '<p class="empty">Preparing import...</p>';
    }

    try {
        const response = await sendJSON('/api/import/google-pay', {
            method: 'POST',
            body: formData
        });
        const jobId = response?.job_id;
        if (!jobId) {
            throw new Error('Import job was not created');
        }

        button.textContent = 'Importing…';
        const job = await waitForGooglePayImport(jobId);
        renderGooglePayImportJob(job);

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

function openEditModal(tx) {
    document.getElementById('editId').value = tx.id || '';
    document.getElementById('editType').value = tx.type || 'Manual';
    document.getElementById('editVendor').value = tx.vendor || '';
    document.getElementById('editAmount').value = tx.amount || '';
    document.getElementById('editCategory').value = tx.category || '';
    document.getElementById('editDate').value = tx.date_time ? tx.date_time.slice(0, 10) : '';
    document.getElementById('editResult').style.display = 'none';
    const modal = document.getElementById('editModal');
    modal.style.display = 'flex';
}

function closeEditModal() {
    document.getElementById('editModal').style.display = 'none';
}

async function saveEditedTransaction() {
    const id = document.getElementById('editId').value;
    const type = document.getElementById('editType').value;
    const vendor = document.getElementById('editVendor').value.trim();
    const amount = parseFloat(document.getElementById('editAmount').value);
    const rawCategory = document.getElementById('editCategory').value;
    const category = rawCategory === '__new__'
        ? document.getElementById('editCategoryNew')?.value.trim()
        : rawCategory;
    const date = document.getElementById('editDate').value;
    const btn = document.getElementById('editSave');
    const result = document.getElementById('editResult');

    if (!vendor) { alert('Merchant is required.'); return; }
    if (!Number.isFinite(amount) || amount === 0) { alert('Enter a non-zero amount.'); return; }
    if (!date) { alert('Date is required.'); return; }

    btn.disabled = true;
    btn.textContent = 'Saving…';

    try {
        await sendJSON('/api/transactions/update', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id, type, vendor, amount, category, date_time: date })
        });
        result.style.display = 'block';
        result.innerHTML = `<p class="empty" style="color:#16a34a">Saved successfully.</p>`;
        await loadDashboard();
        setTimeout(closeEditModal, 800);
    } catch (err) {
        result.style.display = 'block';
        result.innerHTML = `<p class="empty">Error: ${err.message}</p>`;
    } finally {
        btn.disabled = false;
        btn.textContent = 'Save';
    }
}

async function addManualTransaction() {
    const type = document.getElementById('manualType')?.value || 'Manual';
    const vendor = document.getElementById('manualVendor')?.value.trim();
    const amount = parseFloat(document.getElementById('manualAmount')?.value);
    const rawCategory = document.getElementById('manualCategory')?.value;
    const category = rawCategory === '__new__'
        ? document.getElementById('manualCategoryNew')?.value.trim()
        : rawCategory;
    const date = document.getElementById('manualDate')?.value;
    const btn = document.getElementById('manualSubmit');
    const result = document.getElementById('manualResult');

    if (!vendor) { alert('Please enter a merchant name.'); return; }
    if (!Number.isFinite(amount) || amount === 0) { alert('Please enter a non-zero amount.'); return; }
    if (!date) { alert('Please select a date.'); return; }

    btn.disabled = true;
    btn.textContent = 'Saving…';
    result.style.display = 'none';

    try {
        await sendJSON('/api/transactions/manual', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ type, vendor, amount, category, date_time: date })
        });

        result.style.display = 'block';
        result.innerHTML = `<p class="empty" style="color:#16a34a">Transaction saved: ${vendor} — ${formatCurrency(amount)}</p>`;
        document.getElementById('manualVendor').value = '';
        document.getElementById('manualAmount').value = '';
        document.getElementById('manualCategory').value = '';
        const manualCatNew = document.getElementById('manualCategoryNew');
        if (manualCatNew) { manualCatNew.value = ''; manualCatNew.style.display = 'none'; }
        await loadDashboard();
    } catch (err) {
        result.style.display = 'block';
        result.innerHTML = `<p class="empty">Error: ${err.message}</p>`;
    } finally {
        btn.disabled = false;
        btn.textContent = 'Add Transaction';
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

    const manualSubmit = document.getElementById('manualSubmit');
    if (manualSubmit) {
        manualSubmit.addEventListener('click', addManualTransaction);
    }

    document.getElementById('manualCategory')?.addEventListener('change', (e) => {
        const newInput = document.getElementById('manualCategoryNew');
        if (newInput) newInput.style.display = e.target.value === '__new__' ? 'block' : 'none';
    });
    document.getElementById('editCategory')?.addEventListener('change', (e) => {
        const newInput = document.getElementById('editCategoryNew');
        if (newInput) newInput.style.display = e.target.value === '__new__' ? 'block' : 'none';
    });
    document.getElementById('editModalClose')?.addEventListener('click', closeEditModal);
    document.getElementById('editSave')?.addEventListener('click', saveEditedTransaction);
    document.getElementById('editModal')?.addEventListener('click', (e) => {
        if (e.target === document.getElementById('editModal')) closeEditModal();
    });

    document.addEventListener('click', (e) => {
        const row = e.target.closest('.tx-row');
        if (!row) return;
        const id = row.getAttribute('data-id');
        const tx = id && dashboardState.txMap[id];
        if (tx) openEditModal(tx);
    });

    // default date range to current month
    const today = new Date();
    const firstOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
    const fmt = d => d.toISOString().slice(0, 10);
    const fromInput = document.getElementById('rangeFrom');
    const toInput = document.getElementById('rangeTo');
    if (fromInput) fromInput.value = fmt(firstOfMonth);
    if (toInput) toInput.value = fmt(today);
    const manualDate = document.getElementById('manualDate');
    if (manualDate) manualDate.value = fmt(today);

    loadDashboard();
    initChat();
});

// ── Chat widget ──────────────────────────────────────────────────────────────

function initChat() {
    const fab      = document.getElementById('chatFab');
    const panel    = document.getElementById('chatPanel');
    const closeBtn = document.getElementById('chatClose');
    const input    = document.getElementById('chatInput');
    const sendBtn  = document.getElementById('chatSend');
    const messages = document.getElementById('chatMessages');

    let history  = [];
    let loading  = false;

    fab.addEventListener('click', () => {
        const open = panel.style.display === 'flex';
        panel.style.display = open ? 'none' : 'flex';
        if (!open) input.focus();
    });

    closeBtn.addEventListener('click', () => { panel.style.display = 'none'; });

    sendBtn.addEventListener('click', sendMessage);
    input.addEventListener('keydown', e => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); sendMessage(); } });

    async function sendMessage() {
        const question = input.value.trim();
        if (!question || loading) return;

        appendMessage('user', question);
        history.push({ role: 'user', content: question });
        input.value = '';
        loading = true;
        sendBtn.disabled = true;

        const thinkingEl = appendMessage('assistant', '…');

        try {
            const data = await sendJSON('/api/chat', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ question, history: history.slice(0, -1) }),
            });
            thinkingEl.textContent = data.answer;
            history.push({ role: 'assistant', content: data.answer });
        } catch (err) {
            thinkingEl.textContent = 'Error: ' + (err.message || 'something went wrong');
            history.pop();
        } finally {
            loading = false;
            sendBtn.disabled = false;
            input.focus();
        }
    }

    function appendMessage(role, text) {
        const el = document.createElement('div');
        el.className = 'chat-msg chat-msg--' + role;
        el.textContent = text;
        messages.appendChild(el);
        messages.scrollTop = messages.scrollHeight;
        return el;
    }
}
