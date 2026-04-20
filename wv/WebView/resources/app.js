function getSystemInfo() {
    ipc.emit('get-system-info', [], function(result) {
        console.log('系统信息:', result);
        const infoHtml = `
            <strong>操作系统:</strong> ${result.platform}<br>
            <strong>架构:</strong> ${result.arch}<br>
            <strong>浏览器ID:</strong> ${result.browserId}<br>
            <strong>Go版本:</strong> ${result.goVersion}
        `;
        document.getElementById('systemInfo').innerHTML = infoHtml;
    });
}

function calculate() {
    const a = parseFloat(document.getElementById('numA').value);
    const b = parseFloat(document.getElementById('numB').value);
    const operator = document.getElementById('operator').value;

    if (isNaN(a) || isNaN(b)) {
        document.getElementById('calcResult').innerHTML =
            '<span style="color: red;">请输入有效的数字</span>';
        return;
    }

    ipc.emit('calculate', [{a: a, b: b, operator: operator}], function(result) {
        console.log('计算结果:', result);
        if (result.error) {
            document.getElementById('calcResult').innerHTML =
                `<span style="color: red;">${result.error}</span>`;
        } else {
            document.getElementById('calcResult').innerHTML =
                `<strong>结果:</strong> ${a} ${operator} ${b} = <span style="color: #48bb78; font-size: 1.2em;">${result.result}</span>`;
        }
    });
}

function sendMessage() {
    const message = document.getElementById('messageInput').value;
    if (!message) {
        document.getElementById('messageResult').innerHTML =
            '<span style="color: red;">请输入消息内容</span>';
        return;
    }

    ipc.emit('show-message', [message], function(result) {
        console.log('消息响应:', result);
        document.getElementById('messageResult').innerHTML =
            `<strong>响应:</strong> ${result.message}`;
    });
}

function getUserList() {
    ipc.emit('get-user-list', [], function(result) {
        console.log('用户列表:', result);

        let userHtml = '';
        result.forEach(function(user) {
            userHtml += `
                <div class="user-item">
                    <span class="user-id">#${user.id}</span>
                    <span class="user-name">${user.name} (${user.age}岁)</span>
                    <span class="user-email">${user.email}</span>
                </div>
            `;
        });

        document.getElementById('userList').innerHTML = userHtml || '<p>暂无用户数据</p>';
    });
}

document.addEventListener('DOMContentLoaded', function() {
    console.log('Energy WebView 应用已加载');
    console.log('环境信息:', JSON.stringify(energy.env));
});
