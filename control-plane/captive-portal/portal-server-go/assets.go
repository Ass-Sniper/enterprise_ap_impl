package main

const (
	// CommonCSS 包含成功和失败页面的通用样式及动画
	CommonCSS = `
    <style>
        @keyframes fadeOut { from { opacity: 1; } to { opacity: 0; } }
        .auto-submit-msg { animation: fadeOut 2s forwards; animation-delay: 2s; }
        .active-scale:active { transform: scale(0.95); }
        .shake { animation: shake 0.5s cubic-bezier(.36,.07,.19,.97) both; }
        @keyframes shake {
            10%, 90% { transform: translate3d(-1px, 0, 0); }
            20%, 80% { transform: translate3d(2px, 0, 0); }
            30%, 50%, 70% { transform: translate3d(-4px, 0, 0); }
            40%, 60% { transform: translate3d(4px, 0, 0); }
        }
    </style>`

	// SuccessJS 处理登录成功后的自动跳转同步
	SuccessJS = `
	<script>
		window.onload = function() {
			const formData = new FormData(document.getElementById('authForm'));
			
			// 在后台异步发送认证请求
			fetch(document.getElementById('authForm').action, {
				method: 'POST',
				body: formData,
				mode: 'no-cors' // 注意：跨域请求通常需要设置此模式
			}).then(() => {
				console.log("后台准入同步完成");
				// 这里可以更新 UI 状态，比如把“正在同步”改为“准入已完成”
				document.getElementById('syncMsg').innerText = "准入状态已实时同步";
			}).catch(err => {
				console.error("同步失败", err);
			});
		};
	</script>`

	// FailureJS 处理登录失败后的倒计时返回
	FailureJS = `
    <script>
        let seconds = 5;
        const countdownEl = document.getElementById('countdown');
        const timer = setInterval(() => {
            seconds--;
            if (countdownEl) countdownEl.innerText = seconds;
            if (seconds <= 0) {
                clearInterval(timer);
                window.location.href = "/portal";
            }
        }, 1000);
    </script>`
)
