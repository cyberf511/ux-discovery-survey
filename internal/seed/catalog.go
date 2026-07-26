package seed

import "surveyapp/internal/model"

// القاعدة الحاكمة لهذا الكتالوج: كل شاشة لها مستخدم واحد.
//
// السؤال يُوجَّه لمن ينفّذ العمل لا لمن يشرف عليه: الحارس يُسأل كيف يسجّل
// نقطة التفتيش، والمشرف يُسأل كيف يتأكد أن الجولة نُفّذت. لا يُسأل المشرف
// عن QR ولا NFC، ولا يُسأل المشرف العام عن المشي في الجولة، ولا يُسأل مدير
// الشركة عن الحضور والبلاغات.
//
// المشترك بين الجميع ثمانية أسئلة فقط، وهي المصممة للمقارنة بين الأدوار.

const (
	// مشترك
	secCommon = "أسئلة عامة"

	// الحارس — التنفيذ
	secShiftStart = "بداية الوردية والحضور"
	secTasks      = "المهام اليومية"
	secPatrol     = "الدوريات"
	secCheck      = "نقاط التفتيش"
	secReport     = "رفع البلاغات"
	secSOS        = "الاستغاثة والطوارئ"
	secMyRequests = "طلباتي الشخصية"
	secNotify     = "الإشعارات"
	secNetwork    = "الشبكة والانقطاع"
	secUsability  = "سهولة استخدام الجوال"

	// المشرف الميداني — إدارة الحراس داخل المواقع
	secAttendWatch = "متابعة الحضور"
	secAssign      = "توزيع الحراس والمناوبات"
	secPatrolWatch = "متابعة الدوريات"
	secReportWatch = "متابعة البلاغات"
	secApprove     = "اعتماد الطلبات"
	secOpsRoom     = "التواصل مع غرفة العمليات"
	secSiteVisit   = "زيارة المواقع"
	secShiftMap    = "خريطة الوردية ومتابعة الأداء"

	// المشرف العام — عدة مشرفين وعدة مواقع
	secSites     = "مراقبة المواقع"
	secMap       = "الخريطة الشاملة"
	secKPI       = "مؤشرات الأداء و SLA"
	secCritical  = "المشاكل الحرجة"
	secResources = "توزيع الموارد وتغطية الغياب"
	secDecision  = "تقارير الأداء واتخاذ القرار"

	// مدير الشركة — معطّل حاليًا، يُفعَّل عند دراسة الويب
	secProfit    = "الأرباح والنمو"
	secContracts = "العقود والعملاء"
	secBilling   = "الفوترة والتحصيل"
	secPayroll   = "الرواتب والتكاليف"
	secQuality   = "جودة الخدمة والمخاطر"
	secExec      = "التقارير التنفيذية"
)

// commonQuestions الأسئلة المشتركة بين كل الأدوار — ثمانية فقط.
// قيمتها في المقارنة: الفجوة بين إجابة الحارس وإجابة المشرف هي المخرج.
func commonQuestions() []model.Question {
	return []model.Question{
		q(secCommon, "صف لي يوم عملك من بداية الوردية إلى نهايتها.", model.KindLongText, all, true),
		q(secCommon, "ما أكثر شيء يضيع وقتك؟", model.KindLongText, all, true),
		q(secCommon, "ما أكثر شيء يسبب لك الإحباط في العمل؟", model.KindLongText, all, false),
		q(secCommon, "ما أكثر شيء تتمنى تحسينه؟", model.KindLongText, all, false),
		q(secCommon, "ما أكثر شاشة أو أداة تستخدمها يوميًا؟", model.KindLongText, all, false),
		q(secCommon, "لو كان لك قرار واحد لتغيير النظام، ماذا ستغيّر؟", model.KindLongText, all, true),
		q(secCommon, "ما الذي يجعلك تستخدم النظام كل يوم؟", model.KindLongText, all, false),
		q(secCommon, "ما الذي يجعلك تترك النظام وتستخدم واتساب بدلًا عنه؟", model.KindLongText, all, false),
	}
}

// ---------- الحارس ----------

// أقسام الحارس مختصرة عمدًا: الحارس يجيب واقفًا وسط وردية، فطول الاستبيان
// نفسه يفسد نتيجته. أُبقي ما يحسم قرار تصميم، وحُذف ما يصف الحال دون أن يغيّر
// شاشة — والمحذوف باقٍ في تاريخ الشيفرة إن احتجناه في موجة ثانية.

func shiftStartQuestions() []model.Question {
	return []model.Question{
		q(secShiftStart, "كيف تسجل حضورك الآن؟", model.KindLongText, guard, false),
		q(secShiftStart, "هل يحدث أن يرفض تسجيل الحضور؟ ولماذا؟", model.KindLongText, guard, false),
		q(secShiftStart, "كيف تعرف أن حضورك تم بنجاح؟", model.KindLongText, guard, false),
		q(secShiftStart, "إذا فشل التسجيل، ماذا تتوقع أن يحدث؟", model.KindLongText, guard, false),
		q(secShiftStart, "ما رأيك في تصوير وجهك لإثبات الحضور؟", model.KindSingleChoice, guard, false,
			"عادي ولا يزعجني", "مقبول لكن أفضّل غيره", "يزعجني", "أرفضه"),
		q(secShiftStart, "ما المعلومات التي يجب أن تعرفها من الوردية السابقة؟", model.KindLongText, guard, false),
	}
}

func taskQuestions() []model.Question {
	return []model.Question{
		q(secTasks, "كيف تعرف ما هي مهامك اليوم؟", model.KindLongText, guard, false),
		q(secTasks, "هل يحدث أن تفوتك مهمة؟ ولماذا؟", model.KindLongText, guard, false),
		q(secTasks, "كيف تعرف أي مهمة أهم من غيرها؟", model.KindLongText, guard, false),
		q(secTasks, "ما الأشياء التي تحفظها في رأسك وتتمنى أن يتذكرها التطبيق بدلًا عنك؟", model.KindLongText, guard, false),
	}
}

func patrolQuestions() []model.Question {
	return []model.Question{
		q(secPatrol, "كيف تعرف أين تذهب في الجولة؟", model.KindLongText, guard, false),
		q(secPatrol, "كم جولة تنفّذ في الوردية الواحدة؟", model.KindSingleChoice, guard, false,
			"جولة أو جولتان", "من ٣ إلى ٥", "من ٦ إلى ١٠", "أكثر من ١٠"),
		q(secPatrol, "ماذا تحتاج أثناء المشي في الجولة؟", model.KindLongText, guard, false),
		q(secPatrol, "هل تمشي والجوال في يدك أم تحفظه في جيبك؟", model.KindSingleChoice, guard, false,
			"في يدي طوال الجولة", "أخرجه عند الحاجة فقط", "في جيبي غالبًا"),
		q(secPatrol, "كيف تسجّل ملاحظاتك أثناء الجولة الآن؟", model.KindSingleChoice, guard, false,
			"دفتر ورقي", "على الجوال", "أحفظها في ذهني", "لا أسجّل"),
		q(secPatrol, "ماذا يحدث لو قاطعتك حالة طارئة في منتصف الجولة؟", model.KindLongText, guard, false),
	}
}

func checkQuestions() []model.Question {
	return []model.Question{
		q(secCheck, "كيف تثبت أنك وصلت للنقطة؟", model.KindLongText, guard, false),
		q(secCheck, "أي وسيلة إثبات وصول تفضّل؟", model.KindSingleChoice, guard, false,
			"QR", "NFC", "GPS", "لا يهم"),
		q(secCheck, "هل سبق أن واجهت مشكلة في إثبات الوصول؟ صفها.", model.KindLongText, guard, false),
		q(secCheck, "ماذا تفعل لو لم تستطع إثبات وصولك لنقطة؟", model.KindLongText, guard, false),
		q(secCheck, "ما أكثر شيء يبطئك عند نقاط التفتيش؟", model.KindLongText, guard, false),
	}
}

func reportQuestions() []model.Question {
	return []model.Question{
		q(secReport, "متى تنشئ بلاغًا؟", model.KindLongText, guard, false),
		q(secReport, "ما أنواع البلاغات الأكثر شيوعًا عندك؟", model.KindLongText, guard, false),
		q(secReport, "كيف تفضّل تسجيل البلاغ؟", model.KindSingleChoice, guard, false,
			"الكتابة", "الصوت", "الصور", "مزيج بينها"),
		q(secReport, "كم تستغرق كتابة البلاغ الواحد الآن؟", model.KindSingleChoice, guard, false,
			"أقل من دقيقة", "من ١ إلى ٣ دقائق", "من ٣ إلى ١٠ دقائق", "أكثر من ١٠ دقائق"),
		q(secReport, "ما الذي يمنعك أحيانًا من رفع بلاغ رغم وجود سبب؟", model.KindLongText, guard, false),
		q(secReport, "ما الذي تحتاج معرفته بعد رفع البلاغ مباشرة؟", model.KindLongText, guard, false),
	}
}

func sosQuestions() []model.Question {
	return []model.Question{
		q(secSOS, "ماذا تفعل الآن عند وقوع خطر مباشر؟", model.KindLongText, guard, false),
		q(secSOS, "كم ثانية تملك قبل أن يصبح طلب النجدة متأخرًا؟", model.KindSingleChoice, guard, false,
			"أقل من ٥ ثوانٍ", "من ٥ إلى ١٥ ثانية", "من ١٥ إلى ٣٠ ثانية", "أكثر"),
		q(secSOS, "هل تستطيع إخراج جوالك وفتحه أثناء الخطر؟", model.KindBoolean, guard, false),
		q(secSOS, "كيف تريد أن تتأكد أن الاستغاثة وصلت فعلًا؟", model.KindLongText, guard, false),
	}
}

func myRequestQuestions() []model.Question {
	return []model.Question{
		q(secMyRequests, "ما الطلبات التي تقدمها غالبًا؟ (إجازة، استئذان، سلفة...)", model.KindLongText, guard, false),
		q(secMyRequests, "كيف تعرف أن طلبك وصل؟", model.KindLongText, guard, false),
		q(secMyRequests, "ماذا تريد أن ترى بعد إرسال الطلب؟", model.KindLongText, guard, false),
		q(secMyRequests, "كيف تتابع راتبك ومستحقاتك اليوم؟", model.KindLongText, guard, false),
	}
}

func notifyQuestions() []model.Question {
	return []model.Question{
		q(secNotify, "ما الإشعارات المهمة بالنسبة لك؟", model.KindLongText, guard, false),
		q(secNotify, "متى تعتبر الإشعار مزعجًا؟", model.KindLongText, guard, false),
		q(secNotify, "ما الذي يجب أن يظهر في الإشعار دون الحاجة لفتح التطبيق؟", model.KindLongText, guard, false),
		q(secNotify, "هل جوالك على الصامت أثناء العمل؟", model.KindSingleChoice, guard, false,
			"صامت تمامًا", "اهتزاز فقط", "صوت مرتفع", "يختلف"),
	}
}

func networkQuestions() []model.Question {
	return []model.Question{
		q(secNetwork, "في أي أماكن من موقعك تنقطع الشبكة أو تضعف؟", model.KindLongText, guard, false),
		q(secNetwork, "ماذا تفعل الآن لو احتجت تسجيل شيء والشبكة مقطوعة؟", model.KindLongText, guard, false),
		q(secNetwork, "ما الذي يجب أن يعمل بلا إنترنت مهما كان؟", model.KindLongText, guard, true),
		q(secNetwork, "ما الذي يطمئنك أن ما سجّلته حُفظ ولم يضِع؟", model.KindLongText, guard, false),
	}
}

func usabilityQuestions() []model.Question {
	return []model.Question{
		q(secUsability, "هل تستخدم الجوال بيد واحدة أم بكلتا اليدين؟", model.KindSingleChoice, guard, false,
			"بيد واحدة", "بكلتا اليدين", "يختلف حسب الموقف"),
		q(secUsability, "في أي وضع تكون غالبًا أثناء استخدام الجوال؟", model.KindMultiChoice, guard, false,
			"واقف", "أمشي", "داخل السيارة", "جالس"),
		q(secUsability, "كيف ترى شاشة الجوال تحت شمس الظهيرة؟", model.KindSingleChoice, guard, false,
			"واضحة", "أقرأها بصعوبة", "لا أرى شيئًا تقريبًا"),
		q(secUsability, "هل حجم الخط في التطبيقات مناسب لك؟", model.KindSingleChoice, guard, false,
			"مناسب", "صغير أحتاج تكبيره", "كبير زيادة"),
		q(secUsability, "كم ثانية تعتبر انتظارًا مقبولًا بعد الضغط على زر؟", model.KindSingleChoice, guard, false,
			"أقل من ثانية", "من ١ إلى ٣ ثوانٍ", "من ٣ إلى ٥ ثوانٍ", "أكثر من ٥ ثوانٍ"),
		q(secUsability, "عندما تفتح التطبيق، ما أول معلومة تريد رؤيتها؟", model.KindLongText, guard, true),
	}
}

// ---------- المشرف الميداني ----------

func attendWatchQuestions() []model.Question {
	return []model.Question{
		q(secAttendWatch, "كم حارسًا تحت إشرافك؟", model.KindSingleChoice, supervisor, false,
			"أقل من ١٠", "من ١٠ إلى ٣٠", "من ٣٠ إلى ٦٠", "أكثر من ٦٠"),
		q(secAttendWatch, "كيف تعرف الآن أن كل حارس في موقعه؟", model.KindLongText, supervisor, false),
		q(secAttendWatch, "متى تكتشف الغياب عادة؟", model.KindSingleChoice, supervisor, false,
			"فور بداية الوردية", "خلال الساعة الأولى", "بعد ساعات", "بعد انتهاء الوردية"),
		q(secAttendWatch, "كم من وقتك اليومي يذهب في متابعة الحضور؟", model.KindSingleChoice, supervisor, false,
			"أقل من نصف ساعة", "من ٣٠ إلى ٦٠ دقيقة", "من ١ إلى ٣ ساعات", "أكثر"),
		q(secAttendWatch, "ماذا تفعل عند غياب مفاجئ؟", model.KindLongText, supervisor, false),
		q(secAttendWatch, "كيف تعالج خطأ في تسجيل حضور حارس؟", model.KindLongText, supervisor, false),
		q(secAttendWatch, "ما الذي تحتاج رؤيته في شاشة الحضور بنظرة واحدة؟", model.KindLongText, supervisor, true),
		q(secAttendWatch, "هل تحتاج تنبيهًا فوريًا عند تأخر حارس؟ وبعد كم دقيقة؟", model.KindLongText, supervisor, false),
	}
}

func assignQuestions() []model.Question {
	return []model.Question{
		q(secAssign, "كيف توزّع الحراس على المواقع والورديات؟", model.KindLongText, supervisor, false),
		q(secAssign, "كم مرة تعدّل التوزيع في الأسبوع؟", model.KindSingleChoice, supervisor, false,
			"نادرًا", "مرة أو مرتين", "يوميًا تقريبًا", "عدة مرات يوميًا"),
		q(secAssign, "ما الذي يجبرك على إعادة توزيع المناوبات؟", model.KindLongText, supervisor, false),
		q(secAssign, "كم يستغرق إعداد جدول المناوبات؟", model.KindSingleChoice, supervisor, false,
			"أقل من ساعة", "من ١ إلى ٣ ساعات", "يوم عمل", "أكثر"),
		q(secAssign, "ما الذي تراعيه عند اختيار حارس لموقع معيّن؟", model.KindLongText, supervisor, false),
		q(secAssign, "كيف تبلّغ الحارس بتغيير مناوبته؟", model.KindLongText, supervisor, false),
		q(secAssign, "كيف تتأكد أن الحارس علم بالتغيير؟", model.KindLongText, supervisor, false),
		q(secAssign, "هل تحتاج تعديل الجدول من الجوال وأنت في الميدان؟", model.KindBoolean, supervisor, false),
		q(secAssign, "ما الخطأ الأكثر تكرارًا في التوزيع؟", model.KindLongText, supervisor, false),
	}
}

func patrolWatchQuestions() []model.Question {
	return []model.Question{
		q(secPatrolWatch, "كيف تتأكد أن الجولات نُفّذت فعلًا لا شكليًا؟", model.KindLongText, supervisor, false),
		q(secPatrolWatch, "ما العلامة التي تدل على جولة غير حقيقية؟", model.KindLongText, supervisor, false),
		q(secPatrolWatch, "متى تعرف أن جولة لم تُنفَّذ؟", model.KindSingleChoice, supervisor, false,
			"فورًا", "خلال ساعة", "في نهاية الوردية", "في اليوم التالي"),
		q(secPatrolWatch, "ماذا تفعل عند اكتشاف جولة ناقصة؟", model.KindLongText, supervisor, false),
		q(secPatrolWatch, "من يحدد مسارات الجولات وتوقيتها؟", model.KindLongText, supervisor, false),
		q(secPatrolWatch, "كم جولة تراجعها في اليوم؟", model.KindSingleChoice, supervisor, false,
			"أقل من ١٠", "من ١٠ إلى ٣٠", "من ٣٠ إلى ١٠٠", "أكثر من ١٠٠"),
		q(secPatrolWatch, "ما الذي تحتاج رؤيته في تقرير الجولة؟", model.KindLongText, supervisor, false),
		q(secPatrolWatch, "هل تراجع كل جولة أم الاستثناءات فقط؟", model.KindSingleChoice, supervisor, false,
			"كل جولة", "الاستثناءات فقط", "عيّنة عشوائية"),
	}
}

func reportWatchQuestions() []model.Question {
	return []model.Question{
		q(secReportWatch, "كم بلاغًا يصلك في اليوم؟", model.KindSingleChoice, supervisor, false,
			"أقل من ٥", "من ٥ إلى ٢٠", "من ٢٠ إلى ٥٠", "أكثر من ٥٠"),
		q(secReportWatch, "كيف تحدد أي بلاغ عاجل؟", model.KindLongText, supervisor, false),
		q(secReportWatch, "ما المعلومات التي تحتاجها لتقرر خلال ثوانٍ؟", model.KindLongText, supervisor, true),
		q(secReportWatch, "ما الذي ينقص البلاغات التي تصلك عادة؟", model.KindLongText, supervisor, false),
		q(secReportWatch, "ماذا تفعل ببلاغ ناقص المعلومات؟", model.KindLongText, supervisor, false),
		q(secReportWatch, "متى تصعّد البلاغ لمن هو أعلى منك؟", model.KindLongText, supervisor, false),
		q(secReportWatch, "كيف تتابع البلاغات المفتوحة حتى تُغلق؟", model.KindLongText, supervisor, false),
		q(secReportWatch, "ما الذي يتراكم عليك ولا تجد وقتًا لمراجعته؟", model.KindLongText, supervisor, false),
	}
}

func approveQuestions() []model.Question {
	return []model.Question{
		q(secApprove, "ما الطلبات التي تعتمدها أو ترفضها؟", model.KindLongText, supervisor, false),
		q(secApprove, "كم طلبًا تراجع في اليوم؟", model.KindSingleChoice, supervisor, false,
			"أقل من ٥", "من ٥ إلى ٢٠", "من ٢٠ إلى ٥٠", "أكثر من ٥٠"),
		q(secApprove, "على أي أساس توافق أو ترفض؟", model.KindLongText, supervisor, false),
		q(secApprove, "ما المعلومات التي يجب أن تراها قبل الاعتماد؟", model.KindLongText, supervisor, false),
		q(secApprove, "هل تعتمد طلبات كثيرة دفعة واحدة؟", model.KindBoolean, supervisor, false),
		q(secApprove, "ماذا يحدث لو اعتمدت طلبًا بالخطأ؟", model.KindLongText, supervisor, false),
		q(secApprove, "ما الطلب الذي يتأخر عندك أكثر من غيره؟ ولماذا؟", model.KindLongText, supervisor, false),
		q(secApprove, "هل تعتمد الطلبات من الجوال أم تنتظر المكتب؟", model.KindSingleChoice, supervisor, false,
			"من الجوال دائمًا", "من الجوال للعاجل فقط", "أنتظر المكتب"),
	}
}

func opsRoomQuestions() []model.Question {
	return []model.Question{
		q(secOpsRoom, "متى تتواصل مع غرفة العمليات؟", model.KindLongText, supervisor, false),
		q(secOpsRoom, "ما الوسيلة المستخدمة الآن؟", model.KindMultiChoice, supervisor, false,
			"اتصال هاتفي", "واتساب", "جهاز لاسلكي", "رسائل نصية"),
		q(secOpsRoom, "ما المعلومة التي تطلبها منهم باستمرار؟", model.KindLongText, supervisor, false),
		q(secOpsRoom, "ما المعلومة التي يطلبونها منك باستمرار؟", model.KindLongText, supervisor, false),
		q(secOpsRoom, "كم تنتظر ردًا منهم عادة؟", model.KindSingleChoice, supervisor, false,
			"دقائق", "نصف ساعة", "ساعات", "لا يأتي رد أحيانًا"),
		q(secOpsRoom, "ما الذي يضيع بين المناوبات في التواصل؟", model.KindLongText, supervisor, false),
		q(secOpsRoom, "ما مشاكل مجموعات الواتساب في العمل؟", model.KindLongText, supervisor, false),
		q(secOpsRoom, "ما الذي يجب أن يبقى مسجّلًا من المحادثات؟", model.KindLongText, supervisor, false),
	}
}

func siteVisitQuestions() []model.Question {
	return []model.Question{
		q(secSiteVisit, "كم موقعًا تزور في اليوم؟", model.KindSingleChoice, supervisor, false,
			"موقع واحد", "من ٢ إلى ٤", "من ٥ إلى ١٠", "أكثر من ١٠"),
		q(secSiteVisit, "على أي أساس تختار الموقع الذي تزوره؟", model.KindLongText, supervisor, false),
		q(secSiteVisit, "ماذا تفحص عند زيارة الموقع؟", model.KindLongText, supervisor, false),
		q(secSiteVisit, "كيف تسجّل ملاحظاتك بعد الزيارة؟", model.KindLongText, supervisor, false),
		q(secSiteVisit, "هل تُوثَّق الزيارة نفسها؟ كيف؟", model.KindLongText, supervisor, false),
		q(secSiteVisit, "كم من وقتك تقضيه في التنقّل بين المواقع؟", model.KindSingleChoice, supervisor, false,
			"أقل من ساعة", "من ١ إلى ٣ ساعات", "من ٣ إلى ٥ ساعات", "أكثر"),
		q(secSiteVisit, "ما القرار الذي تحتاج اتخاذه بسرعة وأنت في الطريق؟", model.KindLongText, supervisor, false),
		q(secSiteVisit, "هل تستخدم التطبيق داخل السيارة؟ كيف؟", model.KindLongText, supervisor, false),
	}
}

func shiftMapQuestions() []model.Question {
	return []model.Question{
		q(secShiftMap, "هل تحتاج رؤية حراس وردِيّتك على خريطة مباشرة؟", model.KindBoolean, supervisor, false),
		q(secShiftMap, "ماذا تفعل بالمعلومة التي تراها على الخريطة؟", model.KindLongText, supervisor, false),
		q(secShiftMap, "كم دقيقة تأخير في تحديث الموقع تعتبرها مقبولة؟", model.KindSingleChoice, supervisor, false,
			"لحظي", "أقل من دقيقة", "من ١ إلى ٥ دقائق", "لا يهم"),
		q(secShiftMap, "هل تحتاج قائمة بدل الخريطة أحيانًا؟ متى؟", model.KindLongText, supervisor, false),
		q(secShiftMap, "كيف تقيس أداء الحارس اليوم؟", model.KindLongText, supervisor, false),
		q(secShiftMap, "ما الذي يميّز حارسًا ممتازًا عن غيره في نظرك؟", model.KindLongText, supervisor, false),
		q(secShiftMap, "ما الذي تكتشفه متأخرًا وتتمنى معرفته فورًا؟", model.KindLongText, supervisor, true),
		q(secShiftMap, "ما أكثر شكوى تصلك من الحراس؟", model.KindLongText, supervisor, false),
	}
}

// ---------- المشرف العام ----------

func sitesQuestions() []model.Question {
	return []model.Question{
		q(secSites, "كم موقعًا وكم مشرفًا تحت مسؤوليتك؟", model.KindShortText, area, false),
		q(secSites, "كيف تراقب كل المواقع في وقت واحد؟", model.KindLongText, area, false),
		q(secSites, "ما الذي تفتحه أول شيء في بداية يومك؟", model.KindLongText, area, true),
		q(secSites, "كيف تقارن أداء موقع بموقع آخر؟", model.KindLongText, area, false),
		q(secSites, "ما المؤشر الذي يخبرك أن موقعًا فيه مشكلة؟", model.KindLongText, area, false),
		q(secSites, "كم موقعًا تستطيع متابعته فعليًا في وقت واحد؟", model.KindSingleChoice, area, false,
			"أقل من ٥", "من ٥ إلى ١٥", "من ١٥ إلى ٣٠", "أكثر من ٣٠"),
		q(secSites, "ما الفرق بين موقع «هادئ» وموقع «متعثّر» في بياناتك؟", model.KindLongText, area, false),
		q(secSites, "كيف تتابع أداء المشرفين أنفسهم؟", model.KindLongText, area, false),
	}
}

func mapQuestions() []model.Question {
	return []model.Question{
		q(secMap, "ما الذي تريد رؤيته على الخريطة الشاملة؟", model.KindLongText, area, false),
		q(secMap, "هل تحتاج كل المواقع دفعة واحدة أم موقعًا موقعًا؟", model.KindSingleChoice, area, false,
			"كلها دفعة واحدة", "مجموعة مناطق", "موقعًا موقعًا"),
		q(secMap, "ما الذي يجب أن يلفت نظرك على الخريطة فورًا؟", model.KindLongText, area, false),
		q(secMap, "هل تستخدم الخريطة للمراقبة أم لاتخاذ قرار؟", model.KindSingleChoice, area, false,
			"مراقبة فقط", "اتخاذ قرار", "الاثنان"),
		q(secMap, "ما البديل الذي تحتاجه لو كانت الخريطة بطيئة؟", model.KindLongText, area, false),
		q(secMap, "على أي جهاز تفتح الخريطة عادة؟", model.KindSingleChoice, area, false,
			"جوال", "آيباد", "لابتوب", "شاشة كبيرة"),
	}
}

func kpiQuestions() []model.Question {
	return []model.Question{
		q(secKPI, "ما المؤشرات التي تقيس بها الأداء اليوم؟", model.KindLongText, area, true),
		q(secKPI, "كم مؤشرًا تحتاج رؤيته دفعة واحدة؟", model.KindSingleChoice, area, false,
			"من ٣ إلى ٤", "من ٥ إلى ٨", "أكثر من ٨"),
		q(secKPI, "ما اتفاقيات مستوى الخدمة المتفق عليها مع العملاء؟", model.KindLongText, area, false),
		q(secKPI, "كيف تعرف أنك على وشك خرق اتفاقية مستوى خدمة؟", model.KindLongText, area, false),
		q(secKPI, "ما المؤشر الذي يُسأل عنه في الاجتماعات دائمًا؟", model.KindLongText, area, false),
		q(secKPI, "كل كم تراجع المؤشرات؟", model.KindSingleChoice, area, false,
			"عدة مرات يوميًا", "يوميًا", "أسبوعيًا", "شهريًا"),
		q(secKPI, "ما المؤشر الذي تحسبه يدويًا وتتمنى أن يُحسب تلقائيًا؟", model.KindLongText, area, false),
		q(secKPI, "ما الرقم الذي لا تثق به في الأنظمة الحالية؟ ولماذا؟", model.KindLongText, area, false),
	}
}

func criticalQuestions() []model.Question {
	return []model.Question{
		q(secCritical, "ما الذي تعتبره مشكلة حرجة تستدعي تدخلك أنت؟", model.KindLongText, area, false),
		q(secCritical, "كيف تصلك المشاكل الحرجة الآن؟", model.KindLongText, area, false),
		q(secCritical, "كم دقيقة تأخير في إبلاغك تعتبرها غير مقبولة؟", model.KindSingleChoice, area, false,
			"فوري", "أقل من ٥ دقائق", "من ٥ إلى ١٥ دقيقة", "أقل من ساعة"),
		q(secCritical, "ما أول شيء تفعله عند وصول مشكلة حرجة؟", model.KindLongText, area, false),
		q(secCritical, "ما المعلومات التي تحتاجها فورًا لتتصرف؟", model.KindLongText, area, false),
		q(secCritical, "ما الحادثة التي تمنيت لو عرفتها أبكر؟", model.KindLongText, area, false),
		q(secCritical, "كيف تتابع أن المشكلة عولجت فعلًا؟", model.KindLongText, area, false),
	}
}

func resourceQuestions() []model.Question {
	return []model.Question{
		q(secResources, "كيف تخطّط تغطية الإجازات والغياب؟", model.KindLongText, area, false),
		q(secResources, "ما نسبة الغياب التي تتعامل معها أسبوعيًا؟", model.KindSingleChoice, area, false,
			"أقل من ٥٪", "من ٥ إلى ١٠٪", "من ١٠ إلى ٢٠٪", "أكثر"),
		q(secResources, "من أين تأتي بالبديل عادة؟", model.KindLongText, area, false),
		q(secResources, "كم يستغرق سد نقص مفاجئ في موقع؟", model.KindSingleChoice, area, false,
			"أقل من ساعة", "من ١ إلى ٣ ساعات", "نصف يوم", "أكثر"),
		q(secResources, "ما الذي يكلّفك أكثر: الغياب أم العمل الإضافي؟", model.KindLongText, area, false),
		q(secResources, "كيف توازن بين المواقع في عدد الحراس؟", model.KindLongText, area, false),
		q(secResources, "ما المعلومة التي تحتاجها لتقرر نقل حارس من موقع لآخر؟", model.KindLongText, area, false),
	}
}

func decisionQuestions() []model.Question {
	return []model.Question{
		q(secDecision, "ما القرار الأسبوعي الذي تعتمد فيه على أرقام؟", model.KindLongText, area, false),
		q(secDecision, "ما التقرير الذي تُعدّه وتتمنى أن يخرج تلقائيًا؟", model.KindLongText, area, false),
		q(secDecision, "كم من وقتك الأسبوعي يذهب في إعداد التقارير؟", model.KindSingleChoice, area, false,
			"أقل من ساعة", "من ١ إلى ٣ ساعات", "من ٣ إلى ٨ ساعات", "أكثر"),
		q(secDecision, "ما البيانات التي تجمعها يدويًا حتى اليوم؟", model.KindLongText, area, false),
		q(secDecision, "لمن ترفع تقاريرك؟ وماذا يطلبون فيها؟", model.KindLongText, area, false),
		q(secDecision, "هل تصدّر البيانات إلى إكسل؟ ولماذا؟", model.KindLongText, area, false),
		q(secDecision, "ما القرار الذي أخّرته لأن البيانات لم تكن جاهزة؟", model.KindLongText, area, false),
		q(secDecision, "ما الذي تحتاج تنبيهًا عليه قبل فوات الأوان؟", model.KindLongText, area, false),
	}
}

// ---------- مدير الشركة (دور معطّل حاليًا، يُفعَّل عند دراسة الويب) ----------

func profitQuestions() []model.Question {
	return []model.Question{
		q(secProfit, "ما المؤشرات المالية التي تتابعها أسبوعيًا؟", model.KindLongText, manager, false),
		q(secProfit, "كيف تعرف أن عقدًا معيّنًا مربح أو خاسر؟", model.KindLongText, manager, false),
		q(secProfit, "ما أكبر بند تكلفة عندك؟", model.KindLongText, manager, false),
		q(secProfit, "ما الذي يعيق نمو الشركة اليوم؟", model.KindLongText, manager, false),
		q(secProfit, "كيف تقرر دخول مدينة أو قطاع جديد؟", model.KindLongText, manager, false),
	}
}

func contractQuestions() []model.Question {
	return []model.Question{
		q(secContracts, "كيف تتابع تواريخ انتهاء العقود؟", model.KindLongText, manager, false),
		q(secContracts, "ما الذي يجعل العميل يجدّد أو لا يجدّد؟", model.KindLongText, manager, false),
		q(secContracts, "ما أكثر سبب لشكاوى العملاء؟", model.KindLongText, manager, false),
		q(secContracts, "كيف تثبت للعميل أن الخدمة نُفّذت؟", model.KindLongText, manager, false),
		q(secContracts, "هل يحتاج العميل دخولًا على النظام ليرى موقعه؟", model.KindBoolean, manager, false),
		q(secContracts, "ما الرقم الذي يسألك عنه العميل دائمًا؟", model.KindLongText, manager, false),
	}
}

func billingQuestions() []model.Question {
	return []model.Question{
		q(secBilling, "كيف تُحتسب فاتورة العميل اليوم؟", model.KindLongText, manager, false),
		q(secBilling, "كيف تربط ساعات الحراسة الفعلية بالفوترة؟", model.KindLongText, manager, false),
		q(secBilling, "ما الذي يسبب خلافًا على الفاتورة؟", model.KindLongText, manager, false),
		q(secBilling, "كم يستغرق إصدار فواتير الشهر؟", model.KindSingleChoice, manager, false,
			"ساعات", "يوم أو يومان", "أسبوع", "أكثر"),
		q(secBilling, "كيف تتابع التحصيل والمتأخرات؟", model.KindLongText, manager, false),
	}
}

func payrollQuestions() []model.Question {
	return []model.Question{
		q(secPayroll, "كيف تُحتسب الرواتب اليوم؟", model.KindLongText, manager, false),
		q(secPayroll, "ما الذي يتسبب في أخطاء الرواتب؟", model.KindLongText, manager, false),
		q(secPayroll, "كم يستغرق إعداد مسيّر الرواتب؟", model.KindSingleChoice, manager, false,
			"يوم", "من ٢ إلى ٣ أيام", "أسبوع", "أكثر"),
		q(secPayroll, "كيف تتابع العمل الإضافي وتكلفته؟", model.KindLongText, manager, false),
		q(secPayroll, "ما سبب ترك الموظفين للعمل عندكم؟", model.KindLongText, manager, false),
	}
}

func qualityQuestions() []model.Question {
	return []model.Question{
		q(secQuality, "كيف تقيس جودة الخدمة في موقع؟", model.KindLongText, manager, false),
		q(secQuality, "ما المخاطر التي تقلقك أكثر؟", model.KindLongText, manager, false),
		q(secQuality, "كيف تتابع انتهاء الإقامات والتأمين والرخص؟", model.KindLongText, manager, false),
		q(secQuality, "ما الحادثة التي قد تكلّفك عقدًا كاملًا؟", model.KindLongText, manager, false),
		q(secQuality, "ما الذي يجب أن يصلك أنت شخصيًا فورًا؟", model.KindLongText, manager, false),
	}
}

func execQuestions() []model.Question {
	return []model.Question{
		q(secExec, "ما التقرير الذي ترفعه للملّاك أو مجلس الإدارة؟", model.KindLongText, manager, false),
		q(secExec, "ما الأرقام التي تريدها في شاشة واحدة؟", model.KindLongText, manager, false),
		q(secExec, "على أي جهاز تفتح لوحة الإدارة؟", model.KindMultiChoice, manager, false,
			"لابتوب", "كمبيوتر مكتبي", "آيباد", "جوال"),
		q(secExec, "كم مرة تفتح اللوحة في اليوم؟", model.KindSingleChoice, manager, false,
			"مرة أو مرتين", "من ٣ إلى ٥", "أكثر من ٥", "مفتوحة دائمًا"),
		q(secExec, "ما الذي يزعجك في أنظمة الويب التي استخدمتها؟", model.KindLongText, manager, false),
	}
}
