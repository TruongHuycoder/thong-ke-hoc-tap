document.addEventListener('DOMContentLoaded', () => {
    initRadarChart();
    initAreaChart();
});

function initRadarChart() {
    const canvas = document.getElementById('radarChart');
    if (!canvas) return;
    
    const ctx = canvas.getContext('2d');
    const centerX = canvas.width / 2;
    const centerY = canvas.height / 2;
    const radius = 60;
    const labels = ['Demand', 'Supply', 'Equilibrium', 'Elasticity', 'Utility'];
    window.radarValues = window.radarValues || [0.8, 0.7, 0.9, 0.4, 0.85];
    const values = window.radarValues; // 0 to 1
    const sides = labels.length;

window.updateRadarStats = function(newValues) {
    window.radarValues = newValues;
    initRadarChart();
};

    canvas.width = 250;
    canvas.height = 250;
    const adjustedCenterX = canvas.width / 2;
    const adjustedCenterY = canvas.height / 2;

    // Draw background polygons
    ctx.strokeStyle = '#f1f2f6';
    ctx.lineWidth = 1;
    for (let j = 1; j <= 5; j++) {
        const levelRadius = (radius / 5) * j + 20;
        ctx.beginPath();
        for (let i = 0; i < sides; i++) {
            const angle = (Math.PI * 2 / sides) * i - Math.PI / 2;
            const x = adjustedCenterX + levelRadius * Math.cos(angle);
            const y = adjustedCenterY + levelRadius * Math.sin(angle);
            if (i === 0) ctx.moveTo(x, y);
            else ctx.lineTo(x, y);
        }
        ctx.closePath();
        ctx.stroke();
    }

    // Draw axes
    ctx.beginPath();
    for (let i = 0; i < sides; i++) {
        const angle = (Math.PI * 2 / sides) * i - Math.PI / 2;
        const x = adjustedCenterX + (radius + 20) * Math.cos(angle);
        const y = adjustedCenterY + (radius + 20) * Math.sin(angle);
        ctx.moveTo(adjustedCenterX, adjustedCenterY);
        ctx.lineTo(x, y);
    }
    ctx.stroke();

    // Draw data shape
    ctx.beginPath();
    ctx.fillStyle = 'rgba(155, 223, 211, 0.4)';
    ctx.strokeStyle = '#9bdfd3';
    ctx.lineWidth = 2;
    for (let i = 0; i < sides; i++) {
        const angle = (Math.PI * 2 / sides) * i - Math.PI / 2;
        const valRadius = (radius + 20) * values[i];
        const x = adjustedCenterX + valRadius * Math.cos(angle);
        const y = adjustedCenterY + valRadius * Math.sin(angle);
        if (i === 0) ctx.moveTo(x, y);
        else ctx.lineTo(x, y);
    }
    ctx.closePath();
    ctx.fill();
    ctx.stroke();

    // Draw labels
    ctx.fillStyle = '#636e72';
    ctx.font = '10px Inter';
    ctx.textAlign = 'center';
    for (let i = 0; i < sides; i++) {
        const angle = (Math.PI * 2 / sides) * i - Math.PI / 2;
        const labelRadius = radius + 45;
        const x = adjustedCenterX + labelRadius * Math.cos(angle);
        const y = adjustedCenterY + labelRadius * Math.sin(angle);
        ctx.fillText(labels[i], x, y);
    }
}

function initAreaChart() {
    const canvas = document.getElementById('areaChart');
    if (!canvas) return;
    
    const ctx = canvas.getContext('2d');
    canvas.width = canvas.parentElement.clientWidth;
    canvas.height = 200;
    
    const width = canvas.width;
    const height = canvas.height;
    
    // Gradient for Flow Zone
    const flowGradient = ctx.createLinearGradient(0, 0, 0, height);
    flowGradient.addColorStop(0, 'rgba(155, 223, 211, 0)');
    flowGradient.addColorStop(0.5, 'rgba(155, 223, 211, 0.2)');
    flowGradient.addColorStop(1, 'rgba(155, 223, 211, 0)');

    // Background zones
    // Stress (top left), Boredom (bottom right)
    
    // Draw Optimal Curve
    ctx.beginPath();
    ctx.moveTo(0, height * 0.8);
    ctx.bezierCurveTo(width * 0.3, height * 0.7, width * 0.7, height * 0.3, width, height * 0.2);
    
    ctx.strokeStyle = '#9bdfd3';
    ctx.lineWidth = 3;
    ctx.stroke();

    // Fill "Optimal" region (a band around the curve)
    ctx.beginPath();
    ctx.moveTo(0, height * 0.9);
    ctx.bezierCurveTo(width * 0.3, height * 0.8, width * 0.7, height * 0.4, width, height * 0.3);
    ctx.lineTo(width, height * 0.1);
    ctx.bezierCurveTo(width * 0.7, height * 0.2, width * 0.3, height * 0.6, 0, height * 0.7);
    ctx.closePath();
    ctx.fillStyle = 'rgba(155, 223, 211, 0.15)';
    ctx.fill();

    // Markers and text
    ctx.fillStyle = '#636e72';
    ctx.font = '12px Inter';
    ctx.fillText('Difficulty', 10, 20);
    ctx.fillText('Student Skill', width - 100, height - 10);
    
    ctx.font = 'bold 10px Inter';
    ctx.fillStyle = '#9bdfd3';
    ctx.fillText('OPTIMAL ZONE', width / 2 - 30, height / 2);
}

