// Global variables
let analytics = null;
let charts = {};

// ApexCharts color schemes and themes
const chartColors = {
    primary: ['#667eea', '#764ba2'],
    secondary: ['#f093fb', '#f5576c'],
    success: ['#43e97b', '#38f9d7'],
    warning: ['#ffecd2', '#fcb69f'],
    info: ['#4facfe', '#00f2fe'],
    danger: ['#ff9a9e', '#fecfef'],
    categories: [
        '#667eea', '#764ba2', '#f093fb', '#f5576c', '#4facfe', '#00f2fe',
        '#43e97b', '#38f9d7', '#ffecd2', '#fcb69f', '#a8edea', '#fed6e3',
        '#ff9a9e', '#fecfef', '#fad0c4', '#ffd1ff', '#c2e9fb', '#a1c4fd'
    ]
};

// Chart configuration defaults
const chartDefaults = {
    chart: {
        fontFamily: 'Segoe UI, Tahoma, Geneva, Verdana, sans-serif',
        foreColor: '#2c3e50',
        animations: {
            enabled: true,
            easing: 'easeinout',
            speed: 800,
            animateGradually: {
                enabled: true,
                delay: 150
            },
            dynamicAnimation: {
                enabled: true,
                speed: 350
            }
        },
        toolbar: {
            show: true,
            offsetX: 0,
            offsetY: 0,
            tools: {
                download: true,
                selection: false,
                zoom: false,
                zoomin: false,
                zoomout: false,
                pan: false,
                reset: false
            }
        }
    },
    grid: {
        borderColor: '#e0e6ed',
        strokeDashArray: 5,
    },
    tooltip: {
        theme: 'light',
        style: {
            fontSize: '12px',
        }
    }
};

// Utility functions
function formatCurrency(amount) {
    return new Intl.NumberFormat('en-IN', {
        style: 'currency',
        currency: 'INR'
    }).format(amount);
}

function formatDate(dateString) {
    return new Date(dateString).toLocaleDateString('en-IN', {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
    });
}

function showLoading() {
    document.getElementById('loadingSpinner').style.display = 'flex';
    document.getElementById('mainContent').style.display = 'none';
}

function hideLoading() {
    document.getElementById('loadingSpinner').style.display = 'none';
    document.getElementById('mainContent').style.display = 'block';
}

function showError(message, container = 'mainContent') {
    const errorDiv = document.createElement('div');
    errorDiv.className = 'error-message';
    errorDiv.innerHTML = `
        <div class="error-content">
            <i class="fas fa-exclamation-triangle"></i>
            <h3>Something went wrong</h3>
            <p>${message}</p>
            <button onclick="location.reload()" class="retry-btn">
                <i class="fas fa-refresh"></i> Retry
            </button>
        </div>
    `;
    document.getElementById(container).appendChild(errorDiv);
}

// API functions
async function fetchAnalytics() {
    try {
        const response = await fetch('/analytics');
        const data = await response.json();
        if (data.error) throw new Error(data.error);
        return data.analytics;
    } catch (error) {
        console.error('Error fetching analytics:', error);
        throw error;
    }
}

async function fetchInsights() {
    try {
        const response = await fetch('/insights');
        const data = await response.json();
        if (data.error) throw new Error(data.error);
        return data.insights;
    } catch (error) {
        console.error('Error fetching insights:', error);
        throw error;
    }
}

async function fetchRecommendations() {
    try {
        const response = await fetch('/recommendations');
        const data = await response.json();
        if (data.error) throw new Error(data.error);
        return data.recommendations;
    } catch (error) {
        console.error('Error fetching recommendations:', error);
        throw error;
    }
}

async function fetchScore() {
    try {
        const response = await fetch('/score');
        const data = await response.json();
        if (data.error) throw new Error(data.error);
        return data;
    } catch (error) {
        console.error('Error fetching score:', error);
        throw error;
    }
}

async function fetchPredictions() {
    try {
        const response = await fetch('/predictions');
        const data = await response.json();
        if (data.error) throw new Error(data.error);
        return data.predictions;
    } catch (error) {
        console.error('Error fetching predictions:', error);
        throw error;
    }
}

async function askGemini(question) {
    try {
        const response = await fetch('/ask-gemini', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ question })
        });
        const data = await response.json();
        if (data.error) throw new Error(data.error);
        return data.answer;
    } catch (error) {
        console.error('Error asking Gemini:', error);
        throw error;
    }
}

// ApexCharts creation functions
function createCategoryChart(analytics) {
    if (charts.category) {
        charts.category.destroy();
    }

    const categories = Object.keys(analytics.spending_by_category);
    const amounts = Object.values(analytics.spending_by_category);

    const options = {
        ...chartDefaults,
        series: amounts,
        chart: {
            ...chartDefaults.chart,
            type: 'donut',
            height: 350
        },
        labels: categories,
        colors: chartColors.categories,
        plotOptions: {
            pie: {
                donut: {
                    size: '65%',
                    labels: {
                        show: true,
                        total: {
                            show: true,
                            showAlways: false,
                            label: 'Total',
                            fontSize: '22px',
                            fontWeight: 600,
                            color: '#373d3f',
                            formatter: function (w) {
                                return formatCurrency(w.globals.seriesTotals.reduce((a, b) => a + b, 0));
                            }
                        }
                    }
                }
            }
        },
        dataLabels: {
            enabled: true,
            formatter: function (val, opts) {
                return opts.w.config.series[opts.seriesIndex] > 0 ? val.toFixed(1) + "%" : "";
            }
        },
        legend: {
            position: 'bottom',
            horizontalAlign: 'center',
            floating: false,
            fontSize: '14px',
            markers: {
                width: 12,
                height: 12,
                radius: 6
            }
        },
        tooltip: {
            y: {
                formatter: function (val) {
                    return formatCurrency(val);
                }
            }
        }
    };

    charts.category = new ApexCharts(document.querySelector("#categoryChart"), options);
    charts.category.render();
}

function createMonthlyChart(analytics) {
    if (charts.monthly) {
        charts.monthly.destroy();
    }

    const sortedTrends = analytics.monthly_trends.sort((a, b) => a.month.localeCompare(b.month));
    const months = sortedTrends.map(trend => {
        const [year, month] = trend.month.split('-');
        return new Date(year, month - 1).toLocaleDateString('en-IN', { month: 'short', year: 'numeric' });
    });
    const amounts = sortedTrends.map(trend => trend.amount);
    
    // Calculate trend indicators
    const trendData = amounts.map((amount, index) => {
        if (index === 0) return { value: amount, trend: 'neutral' };
        const prevAmount = amounts[index - 1];
        const change = ((amount - prevAmount) / prevAmount) * 100;
        return {
            value: amount,
            trend: change > 5 ? 'up' : change < -5 ? 'down' : 'neutral',
            change: change
        };
    });

    const options = {
        ...chartDefaults,
        series: [{
            name: 'Monthly Spending',
            data: amounts
        }],
        chart: {
            ...chartDefaults.chart,
            type: 'area',
            height: 380,
            dropShadow: {
                enabled: true,
                color: '#000',
                top: 18,
                left: 7,
                blur: 10,
                opacity: 0.2
            },
            background: 'transparent'
        },
        xaxis: {
            categories: months,
            labels: {
                style: {
                    colors: '#64748b',
                    fontSize: '13px',
                    fontWeight: 500
                },
                rotate: -45
            },
            axisBorder: {
                show: false
            },
            axisTicks: {
                show: false
            },
            crosshairs: {
                show: true,
                width: 1,
                position: 'back',
                opacity: 0.9,
                stroke: {
                    color: '#667eea',
                    width: 1,
                    dashArray: 3
                }
            }
        },
        yaxis: {
            labels: {
                formatter: function (val) {
                    return formatCurrency(val);
                },
                style: {
                    colors: '#64748b',
                    fontSize: '13px'
                }
            },
            axisBorder: {
                show: false
            },
            axisTicks: {
                show: false
            }
        },
        colors: ['#667eea'],
        fill: {
            type: 'gradient',
            gradient: {
                shade: 'light',
                type: 'vertical',
                shadeIntensity: 0.4,
                gradientToColors: ['#764ba2', '#f093fb'],
                inverseColors: false,
                opacityFrom: 0.8,
                opacityTo: 0.1,
                stops: [0, 65, 100]
            }
        },
        stroke: {
            curve: 'smooth',
            width: 4,
            lineCap: 'round'
        },
        markers: {
            size: 8,
            colors: ['#fff'],
            strokeColors: '#667eea',
            strokeWidth: 3,
            hover: {
                size: 12,
                sizeOffset: 2
            },
            discrete: amounts.map((amount, index) => ({
                seriesIndex: 0,
                dataPointIndex: index,
                fillColor: trendData[index].trend === 'up' ? '#10b981' : 
                          trendData[index].trend === 'down' ? '#ef4444' : '#667eea',
                strokeColor: '#fff',
                size: 10,
                shape: 'circle'
            }))
        },
        grid: {
            show: true,
            borderColor: '#e2e8f0',
            strokeDashArray: 3,
            position: 'back',
            xaxis: {
                lines: {
                    show: false
                }
            },
            yaxis: {
                lines: {
                    show: true
                }
            }
        },
        tooltip: {
            theme: 'light',
            style: {
                fontSize: '14px'
            },
            custom: function({series, seriesIndex, dataPointIndex, w}) {
                const amount = series[seriesIndex][dataPointIndex];
                const month = months[dataPointIndex];
                const trend = trendData[dataPointIndex];
                
                let trendIcon = '';
                let trendColor = '';
                let trendText = '';
                
                if (trend.trend === 'up') {
                    trendIcon = '📈';
                    trendColor = '#10b981';
                    trendText = `+${trend.change.toFixed(1)}%`;
                } else if (trend.trend === 'down') {
                    trendIcon = '📉';
                    trendColor = '#ef4444';
                    trendText = `${trend.change.toFixed(1)}%`;
                } else {
                    trendIcon = '➖';
                    trendColor = '#6b7280';
                    trendText = 'Stable';
                }
                
                return `
                    <div style="padding: 15px; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); 
                                border-radius: 12px; color: white; box-shadow: 0 10px 25px rgba(0,0,0,0.15);">
                        <div style="font-size: 16px; font-weight: 600; margin-bottom: 8px;">
                            ${month} ${trendIcon}
                        </div>
                        <div style="font-size: 18px; font-weight: 700; margin-bottom: 5px;">
                            ${formatCurrency(amount)}
                        </div>
                        <div style="font-size: 12px; color: ${trendColor}; font-weight: 500;">
                            ${dataPointIndex > 0 ? `${trendText} vs last month` : 'Starting month'}
                        </div>
                    </div>
                `;
            }
        },
        dataLabels: {
            enabled: false
        },
        legend: {
            show: false
        }
    };

    charts.monthly = new ApexCharts(document.querySelector("#monthlyChart"), options);
    charts.monthly.render();
}

function createVendorChart(analytics) {
    if (charts.vendor) {
        charts.vendor.destroy();
    }

    const topVendors = analytics.top_vendors.slice(0, 10);
    const vendors = topVendors.map(vendor => vendor.vendor);
    const amounts = topVendors.map(vendor => vendor.amount);
    
    // Calculate total spending for percentages
    const totalSpending = amounts.reduce((sum, amount) => sum + amount, 0);
    const percentages = amounts.map(amount => (amount / totalSpending) * 100);
    
    // Format vendor names for better readability
    const formattedVendors = vendors.map(vendor => {
        // Capitalize first letter of each word and limit length
        const formatted = vendor.split(' ')
            .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
            .join(' ');
        return formatted.length > 15 ? formatted.substring(0, 15) + '...' : formatted;
    });
    
    // Create dynamic colors based on spending rank
    const dynamicColors = amounts.map((amount, index) => {
        if (index === 0) return '#667eea'; // Top spender - primary blue
        if (index === 1) return '#764ba2'; // Second - purple
        if (index === 2) return '#f093fb'; // Third - pink
        if (index <= 4) return '#4facfe'; // Top 5 - light blue
        return '#a8edea'; // Others - light teal
    });

    const options = {
        ...chartDefaults,
        series: [{
            name: 'Spending Amount',
            data: amounts
        }],
        chart: {
            ...chartDefaults.chart,
            type: 'bar',
            height: 500,
            dropShadow: {
                enabled: true,
                color: '#000',
                top: 0,
                left: 0,
                blur: 10,
                opacity: 0.1
            }
        },
        plotOptions: {
            bar: {
                borderRadius: 6,
                horizontal: true,
                barHeight: '60%',
                distributed: true,
                dataLabels: {
                    position: 'center'
                }
            }
        },
        xaxis: {
            categories: formattedVendors,
            labels: {
                formatter: function (val) {
                    // Format currency values more compactly
                    if (val >= 100000) {
                        return '₹' + (val / 100000).toFixed(1) + 'L';
                    } else if (val >= 1000) {
                        return '₹' + (val / 1000).toFixed(1) + 'K';
                    }
                    return formatCurrency(val);
                },
                style: {
                    colors: '#64748b',
                    fontSize: '12px'
                }
            },
            axisBorder: {
                show: false
            },
            axisTicks: {
                show: false
            }
        },
        yaxis: {
            labels: {
                style: {
                    colors: '#1f2937',
                    fontSize: '13px',
                    fontWeight: 600
                },
                maxWidth: 120
            }
        },
        colors: dynamicColors,
        fill: {
            type: 'gradient',
            gradient: {
                shade: 'light',
                type: 'horizontal',
                shadeIntensity: 0.4,
                gradientToColors: dynamicColors.map(color => color + '80'),
                inverseColors: false,
                opacityFrom: 0.9,
                opacityTo: 0.7,
                stops: [0, 100]
            }
        },
        grid: {
            show: true,
            borderColor: '#e5e7eb',
            strokeDashArray: 3,
            xaxis: {
                lines: {
                    show: true
                }
            },
            yaxis: {
                lines: {
                    show: false
                }
            },
            padding: {
                top: 0,
                right: 30,
                bottom: 0,
                left: 10
            }
        },
        dataLabels: {
            enabled: true,
            formatter: function (val, opts) {
                const percentage = percentages[opts.dataPointIndex];
                // Show only percentage inside bars for cleaner look
                return `${percentage.toFixed(1)}%`;
            },
            offsetX: 0,
            style: {
                fontSize: '12px',
                colors: ['#ffffff'],
                fontWeight: 700
            },
            dropShadow: {
                enabled: true,
                color: '#000',
                top: 1,
                left: 1,
                blur: 2,
                opacity: 0.5
            }
        },
        tooltip: {
            theme: 'light',
            style: {
                fontSize: '14px'
            },
            fixed: {
                enabled: true,
                position: 'topRight',
                offsetX: -20,
                offsetY: 20
            },
            custom: function({series, seriesIndex, dataPointIndex, w}) {
                const amount = series[seriesIndex][dataPointIndex];
                const vendor = vendors[dataPointIndex];
                const percentage = percentages[dataPointIndex];
                const rank = dataPointIndex + 1;
                
                // Get vendor category emoji
                const getVendorIcon = (vendorName) => {
                    const name = vendorName.toLowerCase();
                    if (name.includes('zomato') || name.includes('swiggy') || name.includes('uber eats')) return '🍕';
                    if (name.includes('uber') || name.includes('ola') || name.includes('metro')) return '🚗';
                    if (name.includes('amazon') || name.includes('flipkart') || name.includes('shopping')) return '🛒';
                    if (name.includes('netflix') || name.includes('spotify') || name.includes('youtube')) return '🎬';
                    if (name.includes('paytm') || name.includes('gpay') || name.includes('phonepe')) return '💳';
                    if (name.includes('electricity') || name.includes('gas') || name.includes('water')) return '💡';
                    return '🏪';
                };
                
                const icon = getVendorIcon(vendor);
                const rankEmoji = rank === 1 ? '🥇' : rank === 2 ? '🥈' : rank === 3 ? '🥉' : `#${rank}`;
                
                return `
                    <div style="padding: 18px; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); 
                                border-radius: 15px; color: white; box-shadow: 0 15px 35px rgba(0,0,0,0.2);
                                min-width: 280px; max-width: 320px; z-index: 9999; position: relative;">
                        <div style="display: flex; align-items: center; margin-bottom: 12px;">
                            <span style="font-size: 28px; margin-right: 12px;">${icon}</span>
                            <div style="flex: 1;">
                                <div style="font-size: 17px; font-weight: 700; margin-bottom: 3px; line-height: 1.2;">
                                    ${vendor}
                                </div>
                                <div style="font-size: 13px; opacity: 0.85; display: flex; align-items: center;">
                                    <span style="margin-right: 8px;">${rankEmoji}</span>
                                    <span>Top Spending Vendor</span>
                                </div>
                            </div>
                        </div>
                        <div style="background: rgba(255,255,255,0.15); padding: 12px; border-radius: 8px; margin-bottom: 10px;">
                            <div style="font-size: 22px; font-weight: 800; margin-bottom: 5px;">
                                ${formatCurrency(amount)}
                            </div>
                            <div style="font-size: 14px; opacity: 0.9;">
                                ${percentage.toFixed(1)}% of your total spending
                            </div>
                        </div>
                        <div style="display: flex; justify-content: space-between; align-items: center; font-size: 12px; opacity: 0.9;">
                            <span style="background: rgba(255,255,255,0.2); padding: 4px 8px; border-radius: 12px;">
                                Rank #${rank}
                            </span>
                            <span>Top ${Math.ceil((rank/topVendors.length)*100)}%</span>
                        </div>
                    </div>
                `;
            }
        },
        legend: {
            show: false
        }
    };

    charts.vendor = new ApexCharts(document.querySelector("#vendorChart"), options);
    charts.vendor.render();
}

function createPredictionChart(predictions) {
    if (charts.prediction) {
        charts.prediction.destroy();
    }

    const categories = Object.keys(predictions);
    const amounts = Object.values(predictions);
    
    // Calculate total predicted spending
    const totalPredicted = amounts.reduce((sum, amount) => sum + amount, 0);
    const percentages = amounts.map(amount => (amount / totalPredicted) * 100);
    
    // Get current month data for comparison (from analytics if available)
    const currentMonthData = analytics ? analytics.spending_by_category : {};
    
    // Calculate prediction confidence and changes
    const predictionData = categories.map((category, index) => {
        const predictedAmount = amounts[index];
        const currentAmount = currentMonthData[category] || 0;
        const change = currentAmount > 0 ? ((predictedAmount - currentAmount) / currentAmount) * 100 : 0;
        
        return {
            category,
            amount: predictedAmount,
            percentage: percentages[index],
            change: change,
            confidence: Math.random() * 20 + 75 // Mock confidence score 75-95%
        };
    });
    
    // Sort by predicted amount (highest first)
    predictionData.sort((a, b) => b.amount - a.amount);
    
    const sortedCategories = predictionData.map(d => d.category);
    const sortedAmounts = predictionData.map(d => d.amount);
    
    // Create dynamic colors based on prediction confidence
    const dynamicColors = predictionData.map((data, index) => {
        if (data.confidence > 90) return '#10b981'; // High confidence - green
        if (data.confidence > 85) return '#3b82f6'; // Medium-high - blue
        if (data.confidence > 80) return '#f59e0b'; // Medium - orange
        return '#ef4444'; // Low confidence - red
    });

    const options = {
        ...chartDefaults,
        series: [{
            name: 'Next Month Prediction',
            data: sortedAmounts
        }],
        chart: {
            ...chartDefaults.chart,
            type: 'bar',
            height: 480,
            dropShadow: {
                enabled: true,
                color: '#000',
                top: 0,
                left: 0,
                blur: 10,
                opacity: 0.1
            }
        },
        plotOptions: {
            bar: {
                borderRadius: 8,
                horizontal: true,
                barHeight: '65%',
                distributed: true,
                dataLabels: {
                    position: 'center'
                }
            }
        },
        xaxis: {
            categories: sortedCategories,
            labels: {
                formatter: function (val) {
                    // Format currency values compactly
                    if (val >= 100000) {
                        return '₹' + (val / 100000).toFixed(1) + 'L';
                    } else if (val >= 1000) {
                        return '₹' + (val / 1000).toFixed(1) + 'K';
                    }
                    return formatCurrency(val);
                },
                style: {
                    colors: '#64748b',
                    fontSize: '12px'
                }
            },
            axisBorder: {
                show: false
            },
            axisTicks: {
                show: false
            },
            title: {
                text: 'Predicted Spending Amount',
                style: {
                    fontSize: '14px',
                    fontWeight: 600,
                    color: '#374151'
                }
            }
        },
        yaxis: {
            labels: {
                style: {
                    colors: '#1f2937',
                    fontSize: '13px',
                    fontWeight: 600
                },
                maxWidth: 100
            },
            title: {
                text: 'Categories',
                style: {
                    fontSize: '14px',
                    fontWeight: 600,
                    color: '#374151'
                }
            }
        },
        colors: dynamicColors,
        fill: {
            type: 'gradient',
            gradient: {
                shade: 'light',
                type: 'horizontal',
                shadeIntensity: 0.4,
                gradientToColors: dynamicColors.map(color => color + '80'),
                inverseColors: false,
                opacityFrom: 0.9,
                opacityTo: 0.7,
                stops: [0, 100]
            }
        },
        grid: {
            show: true,
            borderColor: '#e5e7eb',
            strokeDashArray: 3,
            xaxis: {
                lines: {
                    show: true
                }
            },
            yaxis: {
                lines: {
                    show: false
                }
            },
            padding: {
                top: 0,
                right: 30,
                bottom: 0,
                left: 10
            }
        },
        dataLabels: {
            enabled: true,
            formatter: function (val, opts) {
                const data = predictionData[opts.dataPointIndex];
                return `₹${(val / 1000).toFixed(1)}K`;
            },
            offsetX: 0,
            style: {
                fontSize: '12px',
                colors: ['#ffffff'],
                fontWeight: 700
            },
            dropShadow: {
                enabled: true,
                color: '#000',
                top: 1,
                left: 1,
                blur: 2,
                opacity: 0.5
            }
        },
        tooltip: {
            theme: 'light',
            style: {
                fontSize: '14px'
            },
            fixed: {
                enabled: true,
                position: 'topRight',
                offsetX: -20,
                offsetY: 20
            },
            custom: function({series, seriesIndex, dataPointIndex, w}) {
                const data = predictionData[dataPointIndex];
                const currentAmount = currentMonthData[data.category] || 0;
                
                // Get category icon
                const getCategoryIcon = (category) => {
                    const cat = category.toLowerCase();
                    if (cat.includes('food')) return '🍕';
                    if (cat.includes('transport')) return '🚗';
                    if (cat.includes('shopping')) return '🛒';
                    if (cat.includes('entertainment')) return '🎬';
                    if (cat.includes('health')) return '🏥';
                    if (cat.includes('utilities')) return '💡';
                    if (cat.includes('education')) return '📚';
                    if (cat.includes('investment')) return '💰';
                    return '📊';
                };
                
                const icon = getCategoryIcon(data.category);
                const changeIcon = data.change > 5 ? '📈' : data.change < -5 ? '📉' : '➖';
                const changeColor = data.change > 0 ? '#10b981' : data.change < 0 ? '#ef4444' : '#6b7280';
                const confidenceColor = data.confidence > 85 ? '#10b981' : data.confidence > 80 ? '#f59e0b' : '#ef4444';
                
                // Get next month name
                const nextMonth = new Date();
                nextMonth.setMonth(nextMonth.getMonth() + 1);
                const monthName = nextMonth.toLocaleDateString('en-IN', { month: 'long', year: 'numeric' });
                
                return `
                    <div style="padding: 18px; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); 
                                border-radius: 15px; color: white; box-shadow: 0 15px 35px rgba(0,0,0,0.2);
                                min-width: 300px; max-width: 350px; z-index: 9999; position: relative;">
                        <div style="display: flex; align-items: center; margin-bottom: 12px;">
                            <span style="font-size: 28px; margin-right: 12px;">${icon}</span>
                            <div style="flex: 1;">
                                <div style="font-size: 17px; font-weight: 700; margin-bottom: 3px; line-height: 1.2;">
                                    ${data.category}
                                </div>
                                <div style="font-size: 13px; opacity: 0.85;">
                                    ${monthName} Prediction
                                </div>
                            </div>
                            <div style="text-align: center;">
                                <div style="font-size: 24px;">${changeIcon}</div>
                            </div>
                        </div>
                        
                        <div style="background: rgba(255,255,255,0.15); padding: 12px; border-radius: 8px; margin-bottom: 12px;">
                            <div style="font-size: 22px; font-weight: 800; margin-bottom: 5px;">
                                ${formatCurrency(data.amount)}
                            </div>
                            <div style="font-size: 14px; opacity: 0.9;">
                                ${data.percentage.toFixed(1)}% of predicted spending
                            </div>
                        </div>
                        
                        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 10px; margin-bottom: 10px;">
                            <div style="background: rgba(255,255,255,0.1); padding: 8px; border-radius: 6px; text-align: center;">
                                <div style="font-size: 12px; opacity: 0.8;">Current Month</div>
                                <div style="font-size: 14px; font-weight: 600;">
                                    ${formatCurrency(currentAmount)}
                                </div>
                            </div>
                            <div style="background: rgba(255,255,255,0.1); padding: 8px; border-radius: 6px; text-align: center;">
                                <div style="font-size: 12px; opacity: 0.8;">Change</div>
                                <div style="font-size: 14px; font-weight: 600; color: ${changeColor};">
                                    ${data.change > 0 ? '+' : ''}${data.change.toFixed(1)}%
                                </div>
                            </div>
                        </div>
                        
                        <div style="display: flex; justify-content: space-between; align-items: center; font-size: 12px; opacity: 0.9;">
                            <span style="background: rgba(255,255,255,0.2); padding: 4px 8px; border-radius: 12px;">
                                Confidence: ${data.confidence.toFixed(0)}%
                            </span>
                            <span style="color: ${confidenceColor};">
                                ${data.confidence > 85 ? 'High' : data.confidence > 80 ? 'Medium' : 'Low'} Reliability
                            </span>
                        </div>
                    </div>
                `;
            }
        },
        legend: {
            show: false
        },
        title: {
            text: 'AI-Powered Spending Predictions',
            align: 'left',
            style: {
                fontSize: '16px',
                fontWeight: 600,
                color: '#1f2937'
            }
        },
        subtitle: {
            text: 'Based on your spending patterns and trends',
            align: 'left',
            style: {
                fontSize: '13px',
                color: '#6b7280'
            }
        }
    };

    charts.prediction = new ApexCharts(document.querySelector("#predictionChart"), options);
    charts.prediction.render();
}

function createHeatmapChart(analytics) {
    if (charts.heatmap) {
        charts.heatmap.destroy();
    }

    // Create heatmap data based on spending patterns
    const categories = Object.keys(analytics.spending_by_category);
    const months = analytics.monthly_trends.map(trend => {
        const [year, month] = trend.month.split('-');
        return new Date(year, month - 1).toLocaleDateString('en-IN', { month: 'short' });
    });

    // Generate mock data for heatmap (in real app, you'd calculate this from transactions)
    const heatmapData = categories.map(category => {
        return {
            name: category,
            data: months.map((month, index) => {
                // Generate realistic spending patterns
                const baseAmount = analytics.spending_by_category[category] / months.length;
                const variation = (Math.random() - 0.5) * 0.4; // ±20% variation
                return Math.round(baseAmount * (1 + variation));
            })
        };
    });

    const options = {
        ...chartDefaults,
        series: heatmapData,
        chart: {
            ...chartDefaults.chart,
            type: 'heatmap',
            height: 350
        },
        xaxis: {
            categories: months
        },
        colors: [chartColors.info[0]],
        plotOptions: {
            heatmap: {
                shadeIntensity: 0.5,
                radius: 8,
                useFillColorAsStroke: true,
                colorScale: {
                    ranges: [{
                        from: 0,
                        to: 5000,
                        name: 'Low',
                        color: chartColors.success[0]
                    }, {
                        from: 5001,
                        to: 15000,
                        name: 'Medium',
                        color: chartColors.warning[0]
                    }, {
                        from: 15001,
                        to: 50000,
                        name: 'High',
                        color: chartColors.danger[0]
                    }]
                }
            }
        },
        dataLabels: {
            enabled: false
        },
        tooltip: {
            y: {
                formatter: function (val) {
                    return formatCurrency(val);
                }
            }
        }
    };

    charts.heatmap = new ApexCharts(document.querySelector("#heatmapChart"), options);
    charts.heatmap.render();
}

function createBudgetComparisonChart(analytics, recommendations) {
    if (charts.budgetComparison) {
        charts.budgetComparison.destroy();
    }

    const categories = recommendations.map(rec => rec.category);
    const currentSpending = recommendations.map(rec => rec.current_spending);
    const recommendedBudget = recommendations.map(rec => rec.recommended_budget);

    const options = {
        ...chartDefaults,
        series: [{
            name: 'Current Spending',
            data: currentSpending
        }, {
            name: 'Recommended Budget',
            data: recommendedBudget
        }],
        chart: {
            ...chartDefaults.chart,
            type: 'bar',
            height: 350
        },
        xaxis: {
            categories: categories,
            labels: {
                rotate: -45,
                style: {
                    colors: chartDefaults.chart.foreColor,
                    fontSize: '11px'
                }
            }
        },
        yaxis: {
            labels: {
                formatter: function (val) {
                    return formatCurrency(val);
                }
            }
        },
        colors: [chartColors.danger[0], chartColors.success[0]],
        fill: {
            type: 'gradient',
            gradient: {
                shade: 'light',
                type: 'vertical',
                shadeIntensity: 0.3,
                inverseColors: false,
                opacityFrom: 0.8,
                opacityTo: 0.6
            }
        },
        plotOptions: {
            bar: {
                borderRadius: 4,
                dataLabels: {
                    position: 'top'
                }
            }
        },
        dataLabels: {
            enabled: false
        },
        legend: {
            position: 'top',
            horizontalAlign: 'center'
        },
        tooltip: {
            y: {
                formatter: function (val) {
                    return formatCurrency(val);
                }
            }
        }
    };

    charts.budgetComparison = new ApexCharts(document.querySelector("#budgetComparisonChart"), options);
    charts.budgetComparison.render();
}

function createScoreGaugeChart(score) {
    if (charts.scoreGauge) {
        charts.scoreGauge.destroy();
    }

    const options = {
        series: [score.score],
        chart: {
            type: 'radialBar',
            height: 300,
            fontFamily: 'Segoe UI, Tahoma, Geneva, Verdana, sans-serif'
        },
        plotOptions: {
            radialBar: {
                hollow: {
                    margin: 15,
                    size: '60%',
                    image: undefined,
                    imageWidth: 64,
                    imageHeight: 64,
                    imageClipped: false
                },
                dataLabels: {
                    name: {
                        offsetY: -10,
                        show: true,
                        color: '#888',
                        fontSize: '17px'
                    },
                    value: {
                        formatter: function(val) {
                            return parseInt(val);
                        },
                        color: '#111',
                        fontSize: '36px',
                        show: true,
                    }
                }
            }
        },
        fill: {
            type: 'gradient',
            gradient: {
                shade: 'dark',
                type: 'horizontal',
                shadeIntensity: 0.5,
                gradientToColors: score.score >= 80 ? chartColors.success : 
                                 score.score >= 60 ? chartColors.warning : chartColors.danger,
                inverseColors: false,
                opacityFrom: 1,
                opacityTo: 1,
                stops: [0, 100]
            }
        },
        stroke: {
            lineCap: 'round'
        },
        labels: ['Health Score'],
        colors: score.score >= 80 ? chartColors.success : 
                score.score >= 60 ? chartColors.warning : chartColors.danger
    };

    charts.scoreGauge = new ApexCharts(document.querySelector("#scoreGaugeChart"), options);
    charts.scoreGauge.render();
}

// Content population functions
function populateHeaderStats(analytics, score) {
    document.getElementById('totalSpending').textContent = formatCurrency(analytics.total_spending);
    document.getElementById('healthScore').textContent = `${score.score}/100`;
}

function populateInsights(insights) {
    const container = document.getElementById('insightsContainer');
    container.innerHTML = '';

    if (insights.length === 0) {
        container.innerHTML = `
            <div class="no-insights">
                <i class="fas fa-check-circle"></i>
                <h3>Great job!</h3>
                <p>No major spending concerns detected. Keep up the good work!</p>
            </div>
        `;
        return;
    }

    insights.forEach(insight => {
        const insightCard = document.createElement('div');
        insightCard.className = `insight-card ${insight.type}`;
        
        const iconMap = {
            warning: 'fas fa-exclamation-triangle',
            tip: 'fas fa-lightbulb',
            trend: 'fas fa-chart-line'
        };

        insightCard.innerHTML = `
            <div class="insight-header">
                <i class="insight-icon ${iconMap[insight.type]}"></i>
                <span class="insight-category">${insight.category}</span>
            </div>
            <div class="insight-message">${insight.message}</div>
            <div class="insight-impact">
                <strong>Potential Impact:</strong> ${formatCurrency(insight.impact)}
            </div>
        `;

        container.appendChild(insightCard);
    });
}

function populateRecommendations(recommendations) {
    const container = document.getElementById('recommendationsContainer');
    container.innerHTML = '';

    recommendations.forEach(rec => {
        const recCard = document.createElement('div');
        recCard.className = 'recommendation-card';
        
        recCard.innerHTML = `
            <div class="recommendation-header">
                <span class="recommendation-category">${rec.category}</span>
                <span class="savings-badge">Save ${formatCurrency(rec.potential_savings)}</span>
            </div>
            <div class="recommendation-amounts">
                <div class="amount-item">
                    <div class="amount-label">Current</div>
                    <div class="amount-value">${formatCurrency(rec.current_spending)}</div>
                </div>
                <div class="amount-item">
                    <div class="amount-label">Recommended</div>
                    <div class="amount-value">${formatCurrency(rec.recommended_budget)}</div>
                </div>
            </div>
            <div class="recommendation-justification">
                ${rec.justification}
            </div>
        `;

        container.appendChild(recCard);
    });
}

function populateHealthScore(score) {
    document.getElementById('scoreValue').textContent = score.score;
    document.getElementById('scoreExplanation').textContent = score.explanation;
    
    // Update score circle color based on score
    const scoreCircle = document.getElementById('scoreCircle');
    if (score.score >= 80) {
        scoreCircle.style.background = `conic-gradient(${chartColors.success[0]} 0deg, ${chartColors.success[1]} 360deg)`;
    } else if (score.score >= 60) {
        scoreCircle.style.background = `conic-gradient(${chartColors.warning[0]} 0deg, ${chartColors.warning[1]} 360deg)`;
    } else {
        scoreCircle.style.background = `conic-gradient(${chartColors.danger[0]} 0deg, ${chartColors.danger[1]} 360deg)`;
    }
}

function populateTransactions(analytics) {
    const container = document.getElementById('transactionsContainer');
    container.innerHTML = '';

    if (!analytics.recent_transactions || analytics.recent_transactions.length === 0) {
        container.innerHTML = `
            <div class="no-transactions">
                <i class="fas fa-receipt"></i>
                <h3>No Recent Transactions</h3>
                <p>Once you have some transactions, they'll appear here.</p>
            </div>
        `;
        return;
    }

    // Create transactions list
    const transactionsList = document.createElement('div');
    transactionsList.className = 'transactions-list';

    analytics.recent_transactions.forEach((transaction, index) => {
        const transactionCard = document.createElement('div');
        transactionCard.className = 'transaction-card';
        transactionCard.style.animationDelay = `${index * 0.1}s`;
        
        // Get category icon and color
        const categoryInfo = getCategoryInfo(transaction.Category);
        
        // Format transaction type
        const transactionType = getTransactionType(transaction.Type);
        
        // Calculate relative time
        const relativeTime = getRelativeTime(transaction.DateTime);
        
        transactionCard.innerHTML = `
            <div class="transaction-icon" style="background-color: ${categoryInfo.color};">
                <i class="${categoryInfo.icon}"></i>
            </div>
            <div class="transaction-content">
                <div class="transaction-header">
                    <div class="transaction-vendor-info">
                        <span class="transaction-vendor">${transaction.Vendor || 'Unknown'}</span>
                        <span class="transaction-type">${transactionType}</span>
                    </div>
                    <span class="transaction-amount">${formatCurrency(transaction.Amount)}</span>
                </div>
                <div class="transaction-details">
                    <div class="transaction-meta">
                        <span class="transaction-date">
                            <i class="fas fa-clock"></i>
                            ${relativeTime}
                        </span>
                        <span class="transaction-category" style="background-color: ${categoryInfo.color}20; color: ${categoryInfo.color};">
                            ${transaction.Category || 'Other'}
                        </span>
                    </div>
                </div>
            </div>
        `;

        transactionCard.addEventListener('click', () => {
            showTransactionDetails(transaction);
        });

        transactionsList.appendChild(transactionCard);
    });

    container.appendChild(transactionsList);

    // Add "View All" button if there are more than 10 transactions
    if (analytics.recent_transactions.length >= 10) {
        const viewAllBtn = document.createElement('button');
        viewAllBtn.className = 'view-all-btn';
        viewAllBtn.innerHTML = `
            <i class="fas fa-list"></i>
            View All Transactions
        `;
        viewAllBtn.addEventListener('click', showAllTransactions);
        container.appendChild(viewAllBtn);
    }
}

// Helper function to get category information
function getCategoryInfo(category) {
    const categoryMap = {
        'Food': { icon: 'fas fa-utensils', color: '#e74c3c' },
        'General_food': { icon: 'fas fa-shopping-basket', color: '#e67e22' },
        'Shopping': { icon: 'fas fa-shopping-cart', color: '#9b59b6' },
        'Amazon': { icon: 'fab fa-amazon', color: '#ff9900' },
        'Travel': { icon: 'fas fa-plane', color: '#3498db' },
        'Entertainment': { icon: 'fas fa-film', color: '#e91e63' },
        'Bills': { icon: 'fas fa-file-invoice', color: '#f39c12' },
        'Healthcare': { icon: 'fas fa-heartbeat', color: '#2ecc71' },
        'Transfer': { icon: 'fas fa-exchange-alt', color: '#17a2b8' },
        'Other': { icon: 'fas fa-question-circle', color: '#95a5a6' }
    };
    
    return categoryMap[category] || categoryMap['Other'];
}

// Helper function to get transaction type display name
function getTransactionType(type) {
    const typeMap = {
        'HDFCCreditCard': 'HDFC Credit Card',
        'ICICICreditCard': 'ICICI Credit Card',
        'HDFCBankTransfer': 'HDFC Bank Transfer',
        'CreditCard': 'Credit Card',
        'BankTransfer': 'Bank Transfer'
    };
    
    return typeMap[type] || type;
}

// Helper function to get relative time
function getRelativeTime(dateTime) {
    const now = new Date();
    const transactionDate = new Date(dateTime);
    const diffInSeconds = Math.floor((now - transactionDate) / 1000);
    
    if (diffInSeconds < 60) {
        return 'Just now';
    } else if (diffInSeconds < 3600) {
        const minutes = Math.floor(diffInSeconds / 60);
        return `${minutes} minute${minutes > 1 ? 's' : ''} ago`;
    } else if (diffInSeconds < 86400) {
        const hours = Math.floor(diffInSeconds / 3600);
        return `${hours} hour${hours > 1 ? 's' : ''} ago`;
    } else if (diffInSeconds < 604800) {
        const days = Math.floor(diffInSeconds / 86400);
        return `${days} day${days > 1 ? 's' : ''} ago`;
    } else {
        return formatDate(dateTime);
    }
}

// Function to show transaction details in a modal
function showTransactionDetails(transaction) {
    const modal = document.createElement('div');
    modal.className = 'transaction-modal';
    modal.innerHTML = `
        <div class="modal-content">
            <div class="modal-header">
                <h3>Transaction Details</h3>
                <button class="close-btn" onclick="this.parentElement.parentElement.parentElement.remove()">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="detail-row">
                    <span class="detail-label">Vendor:</span>
                    <span class="detail-value">${transaction.Vendor || 'Unknown'}</span>
                </div>
                <div class="detail-row">
                    <span class="detail-label">Amount:</span>
                    <span class="detail-value">${formatCurrency(transaction.Amount)}</span>
                </div>
                <div class="detail-row">
                    <span class="detail-label">Category:</span>
                    <span class="detail-value">${transaction.Category || 'Other'}</span>
                </div>
                <div class="detail-row">
                    <span class="detail-label">Date:</span>
                    <span class="detail-value">${formatDate(transaction.DateTime)}</span>
                </div>
                <div class="detail-row">
                    <span class="detail-label">Type:</span>
                    <span class="detail-value">${getTransactionType(transaction.Type)}</span>
                </div>
                ${transaction.CardEnding ? `
                <div class="detail-row">
                    <span class="detail-label">Card Ending:</span>
                    <span class="detail-value">****${transaction.CardEnding}</span>
                </div>
                ` : ''}
            </div>
        </div>
    `;
    
    // Close modal when clicking outside
    modal.addEventListener('click', (e) => {
        if (e.target === modal) {
            modal.remove();
        }
    });
    
    document.body.appendChild(modal);
}

// Function to show all transactions
function showAllTransactions() {
    // This could be expanded to show a full transactions page
    alert('Full transactions view - coming soon!');
}

// Chat functionality
function addChatMessage(message, isUser = false) {
    const chatHistory = document.getElementById('chatHistory');
    const messageDiv = document.createElement('div');
    messageDiv.className = `chat-message ${isUser ? 'user-message' : 'ai-message'}`;
    messageDiv.textContent = message;
    chatHistory.appendChild(messageDiv);
    chatHistory.scrollTop = chatHistory.scrollHeight;
}

async function handleQuestion(question) {
    const askBtn = document.getElementById('askBtn');
    const questionInput = document.getElementById('questionInput');
    
    // Disable input during processing
    askBtn.disabled = true;
    askBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i>';
    questionInput.disabled = true;
    
    // Add user message
    addChatMessage(question, true);
    
    try {
        const answer = await askGemini(question);
        addChatMessage(answer, false);
    } catch (error) {
        addChatMessage('Sorry, I encountered an error. Please try again.', false);
        console.error('Error in chat:', error);
    } finally {
        // Re-enable input
        askBtn.disabled = false;
        askBtn.innerHTML = '<i class="fas fa-paper-plane"></i>';
        questionInput.disabled = false;
        questionInput.value = '';
    }
}

// Event listeners
document.addEventListener('DOMContentLoaded', async function() {
    // Initialize the dashboard
    await initializeDashboard();
    
    // Set up chat functionality
    const askBtn = document.getElementById('askBtn');
    const questionInput = document.getElementById('questionInput');
    
    askBtn.addEventListener('click', async function() {
        const question = questionInput.value.trim();
        if (question) {
            await handleQuestion(question);
        }
    });
    
    questionInput.addEventListener('keypress', async function(e) {
        if (e.key === 'Enter') {
            const question = questionInput.value.trim();
            if (question) {
                await handleQuestion(question);
            }
        }
    });
    
    // Set up quick question buttons
    document.querySelectorAll('.quick-btn').forEach(btn => {
        btn.addEventListener('click', async function() {
            const question = this.dataset.question;
            questionInput.value = question;
            await handleQuestion(question);
        });
    });
});

// Main initialization function
async function initializeDashboard() {
    showLoading();
    
    try {
        // Fetch all data in parallel
        const [analyticsData, insights, recommendations, score, predictions] = await Promise.all([
            fetchAnalytics(),
            fetchInsights(),
            fetchRecommendations(),
            fetchScore(),
            fetchPredictions()
        ]);
        
        analytics = analyticsData;
        
        // Populate header stats
        populateHeaderStats(analytics, score);
        
        // Create all charts
        createCategoryChart(analytics);
        createMonthlyChart(analytics);
        createVendorChart(analytics);
        createPredictionChart(predictions);
        createHeatmapChart(analytics);
        createBudgetComparisonChart(analytics, recommendations);
        createScoreGaugeChart(score);
        
        // Populate content sections
        populateInsights(insights);
        populateRecommendations(recommendations);
        populateHealthScore(score);
        populateTransactions(analytics);
        
        hideLoading();
        
    } catch (error) {
        hideLoading();
        console.error('Error initializing dashboard:', error);
        showError('Failed to load dashboard data. Please make sure the server is running and try again.');
    }
}

// Make charts responsive
window.addEventListener('resize', function() {
    Object.values(charts).forEach(chart => {
        if (chart && chart.resize) {
            chart.resize();
        }
    });
}); 