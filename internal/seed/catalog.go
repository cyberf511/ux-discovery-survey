package seed

import "surveyapp/internal/model"

// هذا الكتالوج مبني على رحلات AmnAi الموثّقة (J-GRD-01..06، J-SUP-01..03،
// J-GS-01، J-XROLE-01/02) وعلى فصل المنصّات: الميدان جوال فقط، والإدارة ويب.
// أسئلته موجّهة للتحقق مما تصفه الوثائق بأنه Domain-informed target يحتاج قياسًا:
// تكرار المهام، مدة المكوث، المقاطعة، نوع الجهاز، والكمون المقبول.

const (
	secDay      = "فهم يوم العمل"
	secDevice   = "الجهاز والشبكة"
	secPhysical = "بيئة الاستخدام"
	secAuth     = "الدخول والجلسة"
	secBrief    = "الإحاطة وبداية الوردية"
	secAttend   = "الحضور والانصراف"
	secQueue    = "قائمة المهام"
	secPatrol   = "الدوريات"
	secCheck    = "نقاط التفتيش"
	secReport   = "البلاغات"
	secSOS      = "الطوارئ والاستغاثة"
	secCustody  = "العهد والتجهيزات"
	secRequest  = "الطلبات"
	secDirect   = "التعاميم والإقرارات"
	secComm     = "التواصل والاتصال الصوتي"
	secNotify   = "الإشعارات"
	secOffline  = "العمل بلا إنترنت"
	secHome     = "الشاشة الرئيسية"
	secNav      = "التنقّل داخل التطبيق"
	secInput    = "الإدخال والكتابة"
	secLang     = "اللغة والمصطلحات"
	secError    = "الأخطاء والاسترجاع"
	secPerf     = "الأداء"
	secA11y     = "القراءة وسهولة الوصول"
	secTrust    = "الثقة والخصوصية"
	secTrain    = "التدريب والتبنّي"
	secTeam     = "الإشراف على الفريق"
	secReview   = "المراجعة والاعتماد"
	secGS       = "الإشراف العام والخريطة"
	secWebDaily = "لوحة الويب — الاستخدام اليومي"
	secWebData  = "لوحة الويب — الجداول والتقارير"
	secFinance  = "المالية والفوترة والعقود"
	secHR       = "الموارد البشرية والرواتب"
	secClient   = "العملاء والقرارات"
	secHidden   = "الاحتياجات الخفية"
	secPriority = "الأولويات"
	secFinal    = "أسئلة ختامية"
)

// dayQuestions فهم يوم العمل — الإيقاع العام قبل الدخول في أي مهمة بعينها.
func dayQuestions() []model.Question {
	return []model.Question{
		q(secDay, "صف لي يوم عملك من بداية الوردية إلى نهايتها.", model.KindLongText, all, true),
		q(secDay, "ما أول شيء تفعله عندما تبدأ العمل؟", model.KindLongText, all, false),
		q(secDay, "ما أكثر شيء تكرره خلال اليوم؟", model.KindLongText, all, false),
		q(secDay, "ما أكثر شيء يضيع وقتك؟", model.KindLongText, all, true),
		q(secDay, "متى تشعر بالضغط أثناء العمل؟", model.KindLongText, all, false),
		q(secDay, "متى تحتاج استخدام الجوال أثناء العمل؟", model.KindLongText, all, false),
		q(secDay, "هل تستخدم الجوال بيد واحدة أم بكلتا اليدين؟", model.KindSingleChoice, all, false,
			"بيد واحدة", "بكلتا اليدين", "يختلف حسب الموقف"),
		q(secDay, "في أي وضع تكون غالبًا أثناء استخدام الجوال؟", model.KindMultiChoice, all, false,
			"واقف", "أمشي", "داخل السيارة", "جالس في مكتب"),
		q(secDay, "في أي ظروف تستخدم التطبيق؟", model.KindMultiChoice, all, false,
			"تحت الشمس", "في الليل", "أثناء المطر", "داخل مبنى مغطى"),
		q(secDay, "كم ساعة وردتك؟", model.KindSingleChoice, staff, false,
			"٨ ساعات", "١٢ ساعة", "٢٤ ساعة", "يختلف"),
		q(secDay, "كم مرة تفتح تطبيق العمل في الوردية الواحدة؟", model.KindSingleChoice, all, false,
			"أقل من ٥ مرات", "من ٥ إلى ١٥", "من ١٥ إلى ٣٠", "أكثر من ٣٠"),
		q(secDay, "كم ثانية تقضي في التطبيق في المرة الواحدة عادة؟", model.KindSingleChoice, all, false,
			"أقل من ٥ ثوانٍ", "من ٥ إلى ٣٠ ثانية", "من دقيقة إلى ٣ دقائق", "أكثر من ٣ دقائق"),
		q(secDay, "كم مرة تُقاطَع وأنت في منتصف مهمة على الجوال؟", model.KindSingleChoice, all, false,
			"نادرًا", "مرة أو مرتين بالوردية", "عدة مرات", "باستمرار"),
		q(secDay, "ما الذي يقاطعك عادة؟", model.KindLongText, all, false),
		q(secDay, "في أي وقت من الوردية تكون أكثر انشغالًا؟", model.KindSingleChoice, staff, false,
			"بدايتها", "وسطها", "نهايتها", "متفرق بلا نمط"),
		q(secDay, "هل تعمل بمفردك في الموقع أم مع زملاء؟", model.KindSingleChoice, field, false,
			"بمفردي", "مع زميل واحد", "فريق من ٣ فأكثر", "يختلف"),
	}
}

// deviceQuestions الجهاز والشبكة — قيود مادية تحدد حجم التطبيق وسلوكه.
func deviceQuestions() []model.Question {
	return []model.Question{
		q(secDevice, "الجوال الذي تستخدمه للعمل ملكك أم من الشركة؟", model.KindSingleChoice, staff, false,
			"جوالي الشخصي", "جوال من الشركة", "الاثنان"),
		q(secDevice, "ما نظام جوالك؟", model.KindSingleChoice, staff, false,
			"أندرويد", "آيفون", "لا أعرف"),
		q(secDevice, "كم عمر جوالك تقريبًا؟", model.KindSingleChoice, staff, false,
			"أقل من سنة", "من سنة إلى ٣ سنوات", "أكثر من ٣ سنوات", "لا أعرف"),
		q(secDevice, "هل تواجه مشكلة في مساحة التخزين عند تحميل التطبيقات؟", model.KindBoolean, staff, false),
		q(secDevice, "كيف تكون بطارية جوالك في آخر الوردية؟", model.KindSingleChoice, field, false,
			"ممتازة", "تكفي بالضبط", "تخلص قبل النهاية", "تخلص بسرعة كبيرة"),
		q(secDevice, "هل يتوفر لك شاحن أو بطارية متنقلة في الموقع؟", model.KindBoolean, field, false),
		q(secDevice, "ما مصدر الإنترنت لديك أثناء العمل؟", model.KindSingleChoice, field, false,
			"باقة بياناتي", "واي فاي الموقع", "الاثنان", "ضعيف أو منقطع غالبًا"),
		q(secDevice, "في أي أماكن من موقعك تنقطع الشبكة أو تضعف؟", model.KindLongText, field, false),
		q(secDevice, "كم مرة انقطع عنك الإنترنت أثناء العمل الأسبوع الماضي؟", model.KindSingleChoice, field, false,
			"لم ينقطع", "مرة أو مرتين", "عدة مرات", "يوميًا تقريبًا"),
		q(secDevice, "هل تكفي باقة بياناتك لاستخدام تطبيق العمل طوال الشهر؟", model.KindBoolean, field, false),
		q(secDevice, "هل يزعجك حجم التطبيق عند تحميله أو تحديثه؟", model.KindBoolean, staff, false),
	}
}

// physicalQuestions بيئة الاستخدام — القيود الجسدية لا الزمنية.
func physicalQuestions() []model.Question {
	return []model.Question{
		q(secPhysical, "هل تلبس قفازات أثناء العمل؟", model.KindSingleChoice, field, false,
			"دائمًا", "أحيانًا", "أبدًا"),
		q(secPhysical, "هل تستطيع استخدام شاشة الجوال وأنت لابس القفازات؟", model.KindBoolean, field, false),
		q(secPhysical, "كيف ترى شاشة الجوال تحت شمس الظهيرة؟", model.KindSingleChoice, field, false,
			"واضحة", "أقرأها بصعوبة", "لا أرى شيئًا تقريبًا"),
		q(secPhysical, "في الليل، هل إضاءة الشاشة تزعج عينك أو تكشف موقعك؟", model.KindLongText, field, false),
		q(secPhysical, "هل يمنعك شيء أحيانًا من إخراج الجوال؟ (نظام الموقع، العميل، الأمن)", model.KindLongText, field, false),
		q(secPhysical, "هل تستخدم جهاز لاسلكي أو سماعة مع الجوال؟", model.KindBoolean, field, false),
		q(secPhysical, "ما اليد التي تمسك بها الجوال عادة؟", model.KindSingleChoice, all, false,
			"اليمنى", "اليسرى", "تختلف حسب الموقف"),
		q(secPhysical, "هل يصعب عليك الوصول لأعلى الشاشة بإبهامك؟", model.KindBoolean, all, false),
	}
}

// authQuestions الدخول والجلسة — رحلة J-AUTH-01.
func authQuestions() []model.Question {
	return []model.Question{
		q(secAuth, "كيف تدخل على أنظمة العمل الآن؟", model.KindLongText, all, false),
		q(secAuth, "كم مرة تضطر لإعادة تسجيل الدخول؟", model.KindSingleChoice, all, false,
			"مرة واحدة وينتهي", "كل أسبوع تقريبًا", "كل يوم", "عدة مرات في اليوم"),
		q(secAuth, "هل يزعجك انتظار رمز التحقق على الرسائل؟", model.KindLongText, all, false),
		q(secAuth, "ماذا تفعل لو خرجك النظام فجأة وأنت في منتصف مهمة؟", model.KindLongText, all, false),
		q(secAuth, "ماذا تحتاج لو نسيت كلمة المرور وأنت في الموقع؟", model.KindLongText, all, false),
		q(secAuth, "هل تستخدم بصمة الإصبع أو الوجه لفتح جوالك؟", model.KindBoolean, all, false),
		q(secAuth, "هل يشاركك أحد نفس الجوال في العمل؟", model.KindBoolean, field, false),
	}
}

// briefQuestions الإحاطة وبداية الوردية — رحلة J-GRD-01 خطوة الإحاطة والتسليم.
func briefQuestions() []model.Question {
	return []model.Question{
		q(secBrief, "كيف تستلم الوردية من زميلك السابق؟", model.KindLongText, field, false),
		q(secBrief, "ما المعلومات التي يجب أن تعرفها من الوردية السابقة قبل أن تبدأ؟", model.KindLongText, field, false),
		q(secBrief, "هل يحدث أن تستلم بلا تسليم واضح؟", model.KindBoolean, field, false),
		q(secBrief, "ما الأشياء التي تتسلمها ماديًا؟ (مفاتيح، أجهزة، دفاتر)", model.KindLongText, field, false),
		q(secBrief, "كيف يُوثَّق التسليم الآن؟", model.KindSingleChoice, field, false,
			"دفتر ورقي", "واتساب", "اتصال هاتفي", "لا يُوثَّق"),
		q(secBrief, "ما المشاكل التي تحدث بسبب تسليم ناقص؟", model.KindLongText, field, false),
		q(secBrief, "كم يستغرق التسليم عادة؟", model.KindSingleChoice, field, false,
			"أقل من ٥ دقائق", "من ٥ إلى ١٥ دقيقة", "من ١٥ إلى ٣٠ دقيقة", "أكثر من نصف ساعة"),
		q(secBrief, "ماذا تفعل لو لم يحضر بديلك في نهاية وردتك؟", model.KindLongText, field, false),
		q(secBrief, "لو ظهرت لك شاشة إحاطة قبل بدء الوردية، ما الذي يجب أن تحويه؟", model.KindLongText, field, false),
	}
}

// attendQuestions الحضور والانصراف — رحلة J-GRD-01 والتحقق بالوجه.
func attendQuestions() []model.Question {
	return []model.Question{
		q(secAttend, "كيف تسجل حضورك الآن؟", model.KindLongText, field, false),
		q(secAttend, "ما المشاكل التي تواجهها في تسجيل الحضور؟", model.KindLongText, field, false),
		q(secAttend, "هل يحدث أن يرفض تسجيل الحضور؟", model.KindBoolean, field, false),
		q(secAttend, "كيف تعرف أن حضورك تم بنجاح؟", model.KindLongText, field, false),
		q(secAttend, "إذا فشل التسجيل، ماذا تتوقع أن يحدث؟", model.KindLongText, field, false),
		q(secAttend, "كم تبعد نقطة تسجيل الحضور عن مكان عملك الفعلي؟", model.KindShortText, field, false),
		q(secAttend, "هل سبق أن سُجّل غيابك وأنت حاضر فعلًا؟", model.KindBoolean, field, false),
		q(secAttend, "ما الذي يثبت حضورك أمام الإدارة اليوم؟", model.KindLongText, field, false),
		q(secAttend, "هل تنسى تسجيل الانصراف أحيانًا؟ متى؟", model.KindLongText, field, false),
		q(secAttend, "ما رأيك في تصوير وجهك لإثبات الحضور؟", model.KindSingleChoice, field, false,
			"عادي ولا يزعجني", "مقبول لكن أفضّل غيره", "يزعجني", "أرفضه"),
		q(secAttend, "ما الذي يفشل عادة في التصوير؟ (إضاءة، كمامة، نظارة، كاميرا)", model.KindLongText, field, false),
		q(secAttend, "كم ثانية تعتبرها مقبولة لإتمام تسجيل الحضور كاملًا؟", model.KindSingleChoice, field, false,
			"أقل من ٥ ثوانٍ", "من ٥ إلى ١٥ ثانية", "من ١٥ إلى ٣٠ ثانية", "لا يهم المدة"),
	}
}

// queueQuestions قائمة المهام — رحلة J-GRD-02.
func queueQuestions() []model.Question {
	return []model.Question{
		q(secQueue, "كيف تعرف ما هي مهامك اليوم؟", model.KindLongText, field, false),
		q(secQueue, "كم مهمة توكل إليك في الوردية عادة؟", model.KindSingleChoice, field, false,
			"لا شيء محدد", "من ١ إلى ٣", "من ٤ إلى ١٠", "أكثر من ١٠"),
		q(secQueue, "هل يحدث أن تفوتك مهمة؟ ولماذا؟", model.KindLongText, field, false),
		q(secQueue, "كيف تعرف أي مهمة أهم من غيرها؟", model.KindLongText, field, false),
		q(secQueue, "ما الذي تحتاج معرفته عن المهمة قبل أن تبدأها؟", model.KindLongText, field, false),
		q(secQueue, "هل تحتاج إثبات إنجاز المهمة بصورة أو توقيع؟", model.KindSingleChoice, field, false,
			"صورة", "توقيع", "الاثنان", "لا شيء"),
		q(secQueue, "ماذا تفعل لو لم تستطع إنجاز مهمة موكلة إليك؟", model.KindLongText, field, false),
		q(secQueue, "هل تستخدم قوائم تحقق (شيك ليست) في عملك؟", model.KindBoolean, field, false),
	}
}

// patrolQuestions الدوريات — رحلة J-GRD-03، تنفيذ المسار.
func patrolQuestions() []model.Question {
	return []model.Question{
		q(secPatrol, "كيف تعرف أين تذهب في الجولة؟", model.KindLongText, field, false),
		q(secPatrol, "كيف تعرف أنك وصلت؟", model.KindLongText, field, false),
		q(secPatrol, "كيف تعرف أنك انتهيت من الجولة؟", model.KindLongText, field, false),
		q(secPatrol, "ماذا تحتاج أثناء المشي في الجولة؟", model.KindLongText, field, false),
		q(secPatrol, "هل تحتاج الخريطة طوال الوقت؟", model.KindBoolean, field, false),
		q(secPatrol, "ماذا لو انقطع الإنترنت أثناء الجولة؟", model.KindLongText, field, false),
		q(secPatrol, "ماذا لو ضعت عن المسار؟", model.KindLongText, field, false),
		q(secPatrol, "كم جولة تنفّذ في الوردية الواحدة؟", model.KindSingleChoice, field, false,
			"جولة أو جولتان", "من ٣ إلى ٥", "من ٦ إلى ١٠", "أكثر من ١٠"),
		q(secPatrol, "كم تستغرق الجولة الواحدة؟", model.KindSingleChoice, field, false,
			"أقل من ١٥ دقيقة", "من ١٥ إلى ٣٠ دقيقة", "من ٣٠ إلى ٦٠ دقيقة", "أكثر من ساعة"),
		q(secPatrol, "هل مسار الجولة ثابت أم يتغيّر؟", model.KindSingleChoice, field, false,
			"ثابت دائمًا", "يتغيّر يوميًا", "أنا أقرّره", "يتغيّر حسب الحالة"),
		q(secPatrol, "كيف تسجّل ملاحظاتك أثناء الجولة الآن؟", model.KindSingleChoice, field, false,
			"دفتر ورقي", "على الجوال", "أحفظها في ذهني", "لا أسجّل"),
		q(secPatrol, "ماذا يحدث لو قاطعتك حالة طارئة في منتصف الجولة؟", model.KindLongText, field, false),
		q(secPatrol, "هل تمشي والجوال في يدك أم تحفظه في جيبك؟", model.KindSingleChoice, field, false,
			"في يدي طوال الجولة", "أخرجه عند الحاجة فقط", "في جيبي غالبًا"),
		q(secPatrol, "ما الذي يجعلك تعيد جولة أو تتخطاها؟", model.KindLongText, field, false),
	}
}

// checkQuestions نقاط التفتيش — إثبات الوصول تحديدًا لا الجولة كلها.
func checkQuestions() []model.Question {
	return []model.Question{
		q(secCheck, "كيف تثبت أنك وصلت للنقطة؟", model.KindLongText, field, false),
		q(secCheck, "أي وسيلة إثبات وصول تفضّل؟", model.KindSingleChoice, field, false,
			"QR", "NFC", "GPS", "لا يهم"),
		q(secCheck, "هل سبق أن واجهت مشكلة في إثبات الوصول؟ صفها.", model.KindLongText, field, false),
		q(secCheck, "ما أكثر شيء يبطئك عند نقاط التفتيش؟", model.KindLongText, field, false),
		q(secCheck, "كم نقطة تفتيش في موقعك؟", model.KindSingleChoice, field, false,
			"أقل من ٥", "من ٥ إلى ١٠", "من ١٠ إلى ٢٠", "أكثر من ٢٠"),
		q(secCheck, "أين تقع النقاط غالبًا؟", model.KindMultiChoice, field, false,
			"داخل مبنى", "في العراء", "مواقف سيارات", "أسطح أو أقبية"),
		q(secCheck, "هل تصادف نقاطًا صعبة الوصول أو مغلقة؟ صف ذلك.", model.KindLongText, field, false),
		q(secCheck, "ماذا تفعل لو لم تستطع إثبات وصولك لنقطة؟", model.KindLongText, field, false),
		q(secCheck, "كم ثانية مقبولة لتسجيل النقطة الواحدة؟", model.KindSingleChoice, field, false,
			"أقل من ٣ ثوانٍ", "من ٣ إلى ١٠ ثوانٍ", "أكثر من ١٠ ثوانٍ"),
		q(secCheck, "هل تحتاج كتابة ملاحظة عند كل نقطة أم يكفي الإثبات؟", model.KindSingleChoice, field, false,
			"يكفي الإثبات", "ملاحظة عند الحاجة فقط", "ملاحظة في كل نقطة"),
	}
}

// reportQuestions البلاغات — التوثيق غير العاجل.
func reportQuestions() []model.Question {
	return []model.Question{
		q(secReport, "متى تنشئ بلاغًا؟", model.KindLongText, all, false),
		q(secReport, "ما أنواع البلاغات الأكثر شيوعًا عندك؟", model.KindLongText, all, false),
		q(secReport, "كيف تفضّل تسجيل البلاغ؟", model.KindSingleChoice, all, false,
			"الكتابة", "الصوت", "الصور", "مزيج بينها"),
		q(secReport, "ما المعلومات التي يجب أن تُطلب منك في البلاغ؟", model.KindLongText, all, false),
		q(secReport, "ما المعلومات التي لا ينبغي أن تُطلب منك؟", model.KindLongText, all, false),
		q(secReport, "كم بلاغًا ترفع في الأسبوع تقريبًا؟", model.KindSingleChoice, all, false,
			"لا شيء غالبًا", "من ١ إلى ٣", "من ٤ إلى ١٠", "أكثر من ١٠"),
		q(secReport, "ما الذي يمنعك أحيانًا من رفع بلاغ رغم وجود سبب؟", model.KindLongText, all, false),
		q(secReport, "كم تستغرق كتابة البلاغ الواحد الآن؟", model.KindSingleChoice, all, false,
			"أقل من دقيقة", "من ١ إلى ٣ دقائق", "من ٣ إلى ١٠ دقائق", "أكثر من ١٠ دقائق"),
		q(secReport, "كم صورة ترفقها بالبلاغ عادة؟", model.KindSingleChoice, all, false,
			"لا أرفق صورًا", "صورة واحدة", "من ٢ إلى ٤", "أكثر من ٤"),
		q(secReport, "ما الذي تحتاج معرفته بعد رفع البلاغ مباشرة؟", model.KindLongText, all, false),
		q(secReport, "كم تنتظر ردًا على البلاغ عادة؟", model.KindSingleChoice, all, false,
			"دقائق", "ساعات", "يوم أو أكثر", "لا يأتي رد غالبًا"),
		q(secReport, "هل تعود لمراجعة بلاغاتك القديمة؟ ولماذا؟", model.KindLongText, all, false),
	}
}

// sosQuestions الطوارئ — رحلة J-GRD-04، مسار حرج منفصل عن البلاغ العادي.
func sosQuestions() []model.Question {
	return []model.Question{
		q(secSOS, "صف آخر موقف طارئ مررت به في العمل.", model.KindLongText, field, false),
		q(secSOS, "ماذا تفعل الآن عند وقوع خطر مباشر؟", model.KindLongText, field, false),
		q(secSOS, "كم ثانية تملك قبل أن يصبح طلب النجدة متأخرًا؟", model.KindSingleChoice, field, false,
			"أقل من ٥ ثوانٍ", "من ٥ إلى ١٥ ثانية", "من ١٥ إلى ٣٠ ثانية", "أكثر"),
		q(secSOS, "هل تستطيع إخراج جوالك وفتحه أثناء الخطر؟", model.KindBoolean, field, false),
		q(secSOS, "ما الذي يجب أن يصل للمشرف تلقائيًا عند استغاثتك؟", model.KindLongText, field, false),
		q(secSOS, "كيف تريد أن تتأكد أن الاستغاثة وصلت فعلًا؟", model.KindLongText, field, false),
		q(secSOS, "هل تخشى إطلاق استغاثة بالخطأ؟ وما الذي يطمئنك؟", model.KindLongText, field, false),
		q(secSOS, "هل تحتاج طلب النجدة بلا صوت أو بلا أن يراك أحد؟", model.KindBoolean, field, false),
	}
}

// custodyQuestions العهد — رحلة J-GRD-06، وهي ONLINE_ONLY في التصميم الحالي.
func custodyQuestions() []model.Question {
	return []model.Question{
		q(secCustody, "ما التجهيزات التي في عهدتك؟", model.KindLongText, field, false),
		q(secCustody, "كيف تُسلَّم إليك العهدة وتُستلم منك اليوم؟", model.KindLongText, field, false),
		q(secCustody, "هل حدث خلاف حول عهدة مفقودة أو تالفة؟ صف ما جرى.", model.KindLongText, field, false),
		q(secCustody, "ما الذي يثبت أن العهدة سُلّمت سليمة؟", model.KindLongText, field, false),
		q(secCustody, "هل تحتاج تصوير حالة الجهاز عند الاستلام؟", model.KindBoolean, field, false),
		q(secCustody, "هل يحدث تسليم عهدة في مكان بلا تغطية شبكة؟", model.KindBoolean, field, false),
		q(secCustody, "كم مرة تُجرد العهد في موقعك؟", model.KindSingleChoice, field, false,
			"أسبوعيًا", "شهريًا", "عند التسليم فقط", "لا يوجد جرد منتظم"),
	}
}

// requestQuestions الطلبات والسجلات الشخصية — رحلة J-GRD-05.
func requestQuestions() []model.Question {
	return []model.Question{
		q(secRequest, "ما الطلبات التي تقدمها غالبًا؟ (إجازة، استئذان، سلفة...)", model.KindLongText, staff, false),
		q(secRequest, "ما المعلومات المطلوبة لكل طلب؟", model.KindLongText, staff, false),
		q(secRequest, "كيف تعرف أن الطلب وصل؟", model.KindLongText, staff, false),
		q(secRequest, "ماذا تريد أن ترى بعد إرسال الطلب؟", model.KindLongText, staff, false),
		q(secRequest, "كم طلبًا تقدّم في الشهر تقريبًا؟", model.KindSingleChoice, staff, false,
			"لا شيء غالبًا", "طلب أو طلبان", "من ٣ إلى ٥", "أكثر من ٥"),
		q(secRequest, "كم تستغرق الموافقة عادة؟", model.KindSingleChoice, staff, false,
			"ساعات", "يوم أو يومان", "أسبوع", "أكثر من أسبوع"),
		q(secRequest, "هل تعرف من هو المسؤول عن الموافقة على طلبك؟", model.KindBoolean, staff, false),
		q(secRequest, "هل احتجت مرة إلغاء أو تعديل طلب بعد إرساله؟", model.KindBoolean, staff, false),
		q(secRequest, "كيف تتابع راتبك ومستحقاتك اليوم؟", model.KindLongText, staff, false),
		q(secRequest, "لو كان عليك جزاء أو خصم، ماذا تحتاج أن ترى؟", model.KindLongText, staff, false),
		q(secRequest, "هل قدّمت تظلّمًا من قبل؟ كيف كانت التجربة؟", model.KindLongText, staff, false),
	}
}

// directiveQuestions التعاميم والإقرارات — إقرار القراءة والتوقيع.
func directiveQuestions() []model.Question {
	return []model.Question{
		q(secDirect, "كيف تصلك تعاميم الشركة وتعليماتها؟", model.KindLongText, all, false),
		q(secDirect, "هل تقرأها كاملة؟ ولماذا؟", model.KindLongText, all, false),
		q(secDirect, "كيف تثبت الشركة أنك قرأت التعميم؟", model.KindLongText, all, false),
		q(secDirect, "ما طول التعميم الذي تقرأه فعلًا حتى نهايته؟", model.KindSingleChoice, all, false,
			"سطران أو ثلاثة", "فقرة قصيرة", "صفحة كاملة", "أي طول إن كان مهمًا"),
		q(secDirect, "هل وقّعت إلكترونيًا على مستند من قبل؟", model.KindBoolean, all, false),
		q(secDirect, "ما الذي يقلقك في التوقيع الإلكتروني؟", model.KindLongText, all, false),
		q(secDirect, "ما التعليمات التي تحتاج الرجوع إليها باستمرار؟", model.KindLongText, field, false),
	}
}

// commQuestions التواصل والاتصال الصوتي — رحلة J-XROLE-01.
func commQuestions() []model.Question {
	return []model.Question{
		q(secComm, "مع من تتواصل أكثر أثناء العمل؟", model.KindSingleChoice, all, false,
			"المشرف المباشر", "الزملاء في الموقع", "غرفة العمليات", "العميل"),
		q(secComm, "ما الوسائل المستخدمة اليوم؟", model.KindMultiChoice, all, false,
			"اتصال هاتفي", "واتساب", "جهاز لاسلكي", "رسائل نصية", "وجهًا لوجه"),
		q(secComm, "ما الذي يجعلك تتصل بدل أن تكتب؟", model.KindLongText, all, false),
		q(secComm, "ما مشاكل مجموعات الواتساب في العمل؟", model.KindLongText, all, false),
		q(secComm, "متى تصعّد الموضوع لمن هو أعلى منك؟", model.KindLongText, all, false),
		q(secComm, "كم تنتظر ردًا من المشرف عادة؟", model.KindSingleChoice, field, false,
			"دقائق", "نصف ساعة", "ساعات", "لا يأتي رد أحيانًا"),
		q(secComm, "هل تضيع تعليمات بين الورديات؟", model.KindBoolean, field, false),
		q(secComm, "هل جربت خاصية «اضغط للتحدث»؟ وهل تناسب عملك؟", model.KindLongText, field, false),
		q(secComm, "هل تحتاج التواصل مع زملاء في مواقع أخرى؟", model.KindBoolean, all, false),
		q(secComm, "ما الذي يجب أن يبقى مسجّلًا من المحادثات، وما الذي لا؟", model.KindLongText, all, false),
	}
}

// notifyQuestions الإشعارات.
func notifyQuestions() []model.Question {
	return []model.Question{
		q(secNotify, "ما الإشعارات المهمة بالنسبة لك؟", model.KindLongText, all, false),
		q(secNotify, "متى تعتبر الإشعار مزعجًا؟", model.KindLongText, all, false),
		q(secNotify, "كيف تميّز بين الإشعار العادي والعاجل؟", model.KindLongText, all, false),
		q(secNotify, "ما الذي يجب أن يظهر في الإشعار دون الحاجة لفتح التطبيق؟", model.KindLongText, all, false),
		q(secNotify, "كم إشعارًا في اليوم تعتبره مقبولًا من تطبيق العمل؟", model.KindSingleChoice, all, false,
			"أقل من ٥", "من ٥ إلى ١٥", "من ١٥ إلى ٣٠", "لا يهم إن كانت مفيدة"),
		q(secNotify, "هل جوالك على الصامت أثناء العمل؟", model.KindSingleChoice, all, false,
			"صامت تمامًا", "اهتزاز فقط", "صوت مرتفع", "يختلف"),
		q(secNotify, "هل تفوتك إشعارات مهمة؟ ولماذا؟", model.KindLongText, all, false),
		q(secNotify, "هل تحتاج نغمة مختلفة للحالات العاجلة؟", model.KindBoolean, all, false),
	}
}

// offlineQuestions العمل بلا إنترنت — سياسة الطابور والإيصالات.
func offlineQuestions() []model.Question {
	return []model.Question{
		q(secOffline, "ماذا تفعل الآن لو احتجت تسجيل شيء والشبكة مقطوعة؟", model.KindLongText, field, false),
		q(secOffline, "ما الذي يجب أن يعمل بلا إنترنت مهما كان؟", model.KindLongText, field, true),
		q(secOffline, "هل تثق بأن ما سجّلته بلا إنترنت سيصل لاحقًا؟", model.KindLongText, field, false),
		q(secOffline, "ما الذي يطمئنك أن العمل حُفظ ولم يضِع؟", model.KindLongText, field, false),
		q(secOffline, "هل يزعجك أن يُرسل التطبيق البيانات لاحقًا بلا إذنك؟", model.KindBoolean, field, false),
		q(secOffline, "هل حدث أن سجّلت شيئًا مرتين لأنك لم تتأكد من وصوله؟", model.KindBoolean, field, false),
	}
}

// homeQuestions الشاشة الرئيسية — أولوية المعلومة لا التنقّل.
func homeQuestions() []model.Question {
	return []model.Question{
		q(secHome, "عندما تفتح التطبيق، ما أول معلومة تريد رؤيتها؟", model.KindLongText, all, true),
		q(secHome, "ما أهم ثلاث معلومات بالنسبة لك؟", model.KindLongText, all, false),
		q(secHome, "ما الأشياء التي لا تحتاج رؤيتها كل مرة؟", model.KindLongText, all, false),
		q(secHome, "إذا فتحت التطبيق لمدة خمس ثوانٍ فقط، ماذا يجب أن تعرف؟", model.KindLongText, all, false),
		q(secHome, "ما الحالة أو الرقم الذي تريده كبيرًا وواضحًا في الأعلى؟", model.KindLongText, all, false),
		q(secHome, "كم زرًا رئيسيًا يكفيك في الشاشة الأولى؟", model.KindSingleChoice, all, false,
			"زران", "ثلاثة", "أربعة", "خمسة فأكثر"),
		q(secHome, "هل تحتاج شاشة مختلفة قبل الوردية وأثناءها؟", model.KindBoolean, field, false),
		q(secHome, "ما الشيء الذي تبحث عنه ولا تجده بسرعة في الأنظمة التي جرّبتها؟", model.KindLongText, all, false),
	}
}

// navQuestions التنقّل وبنية التطبيق — بنية المعلومات لا محتواها.
func navQuestions() []model.Question {
	return []model.Question{
		q(secNav, "كم ضغطة تعتبرها مقبولة للوصول إلى أهم مهمة عندك؟", model.KindSingleChoice, all, false,
			"ضغطة واحدة", "ضغطتان", "ثلاث", "لا يهم العدد"),
		q(secNav, "هل تضيع في التطبيقات ذات القوائم الكثيرة؟", model.KindBoolean, all, false),
		q(secNav, "ما الشاشة التي تتمنى أن تكون بضغطة واحدة؟", model.KindLongText, all, false),
		q(secNav, "هل تفضّل صفحة واحدة طويلة أم صفحات مقسّمة؟", model.KindSingleChoice, all, false,
			"صفحة واحدة أنزّل فيها", "صفحات مقسّمة", "لا فرق"),
		q(secNav, "هل تحتاج بحثًا داخل التطبيق؟ وعن ماذا تبحث؟", model.KindLongText, all, false),
		q(secNav, "هل تحتاج اختصارات تصنعها بنفسك؟", model.KindBoolean, all, false),
		q(secNav, "ما التطبيق الذي تعتبره سهلًا وتتمنى أن يشبهه تطبيقنا؟", model.KindShortText, all, false),
		q(secNav, "ما الذي يجعلك تخرج من تطبيق وتتصل بالمشرف بدلًا عنه؟", model.KindLongText, all, false),
	}
}

// inputQuestions الإدخال والكتابة.
func inputQuestions() []model.Question {
	return []model.Question{
		q(secInput, "هل الكتابة بالعربي على جوالك مريحة؟", model.KindBoolean, all, false),
		q(secInput, "هل تستخدم الإملاء الصوتي بدل الكتابة؟", model.KindBoolean, all, false),
		q(secInput, "هل تفضّل الاختيار من قائمة جاهزة بدل الكتابة؟", model.KindBoolean, all, false),
		q(secInput, "ما الحقول التي تتمنى أن تُعبّأ تلقائيًا؟", model.KindLongText, all, false),
		q(secInput, "هل يزعجك تكرار كتابة نفس المعلومة في كل مرة؟", model.KindBoolean, all, false),
		q(secInput, "ما الأخطاء التي تحدث لك عند الكتابة على الجوال؟", model.KindLongText, all, false),
		q(secInput, "ما أطول نص اضطررت لكتابته للعمل؟", model.KindShortText, all, false),
		q(secInput, "هل تكتب وأنت تمشي أو واقف؟", model.KindBoolean, field, false),
	}
}

// langQuestions اللغة والمصطلحات — مادة خام لنصوص الواجهة.
func langQuestions() []model.Question {
	return []model.Question{
		q(secLang, "هل تفضّل التطبيق بالعربي فقط أم عربي وإنجليزي؟", model.KindSingleChoice, all, false,
			"عربي فقط", "عربي مع إنجليزي", "إنجليزي فقط", "لا يهم"),
		q(secLang, "هل توجد مصطلحات في الأنظمة الحالية لا تفهمها؟ اذكرها.", model.KindLongText, all, false),
		q(secLang, "ماذا تسمّي «الجولة» في كلامك اليومي؟", model.KindShortText, field, false),
		q(secLang, "ماذا تسمّي «البلاغ» في كلامك اليومي؟", model.KindShortText, all, false),
		q(secLang, "ماذا تسمّي «العهدة» في كلامك اليومي؟", model.KindShortText, field, false),
		q(secLang, "هل بين زملائك من لغته الأولى ليست العربية؟", model.KindBoolean, all, false),
		q(secLang, "هل تفضّل لغة رسمية أم قريبة من العامية داخل التطبيق؟", model.KindSingleChoice, all, false,
			"رسمية", "قريبة من العامية", "لا يهم"),
	}
}

// errorQuestions الأخطاء والاسترجاع — رحلة J-RECOVERY-01.
func errorQuestions() []model.Question {
	return []model.Question{
		q(secError, "ماذا تفعل لو تعلّق التطبيق في منتصف مهمة؟", model.KindLongText, all, false),
		q(secError, "هل تعيد المحاولة أم تتصل بالمشرف مباشرة؟", model.KindSingleChoice, all, false,
			"أعيد المحاولة", "أتصل بالمشرف", "أتركه وأكمل لاحقًا"),
		q(secError, "ما رسالة الخطأ التي أزعجتك في نظام سابق؟", model.KindLongText, all, false),
		q(secError, "هل تحتاج معرفة سبب الخطأ أم يكفي «حاول مرة أخرى»؟", model.KindSingleChoice, all, false,
			"أحتاج السبب", "يكفي أن أعرف ماذا أفعل", "لا يهم"),
		q(secError, "ماذا تتوقع لو ضغطت إرسال ولم يصل؟", model.KindLongText, all, false),
		q(secError, "هل حدث أن أرسلت الشيء نفسه مرتين بالخطأ؟", model.KindBoolean, all, false),
		q(secError, "هل حدث أن سقط جوالك أو انطفأ أثناء مهمة؟ ماذا جرى بعدها؟", model.KindLongText, field, false),
		q(secError, "ما الذي يجب ألّا يضيع أبدًا مهما حدث؟", model.KindLongText, all, false),
	}
}

// perfQuestions الأداء والانتظار.
func perfQuestions() []model.Question {
	return []model.Question{
		q(secPerf, "كم ثانية تعتبر انتظارًا مقبولًا؟", model.KindSingleChoice, all, false,
			"أقل من ثانية", "من ١ إلى ٣ ثوانٍ", "من ٣ إلى ٥ ثوانٍ", "أكثر من ٥ ثوانٍ"),
		q(secPerf, "إذا كان الإنترنت ضعيفًا، ماذا تتوقع من التطبيق؟", model.KindLongText, all, false),
		q(secPerf, "هل تفضّل أن يعمل التطبيق بدون إنترنت؟", model.KindBoolean, all, false),
		q(secPerf, "ما أسوأ شيء يمكن أن يحدث أثناء العمل؟", model.KindLongText, all, false),
		q(secPerf, "كم ثانية مقبولة لفتح التطبيق أول مرة في اليوم؟", model.KindSingleChoice, all, false,
			"أقل من ثانيتين", "من ٢ إلى ٥ ثوانٍ", "من ٥ إلى ١٠ ثوانٍ", "لا يهم"),
		q(secPerf, "ما الذي يجعلك تصف تطبيقًا بأنه «ثقيل»؟", model.KindLongText, all, false),
		q(secPerf, "هل تُغلق تطبيقات لتوفير البطارية؟", model.KindBoolean, all, false),
		q(secPerf, "ما رأيك أن يعمل التطبيق في الخلفية طوال الوردية؟", model.KindLongText, field, false),
	}
}

// a11yQuestions القراءة وسهولة الوصول.
func a11yQuestions() []model.Question {
	return []model.Question{
		q(secA11y, "هل حجم الخط في التطبيقات مناسب لك؟", model.KindSingleChoice, all, false,
			"مناسب", "صغير أحتاج تكبيره", "كبير زيادة"),
		q(secA11y, "هل كبّرت حجم الخط في إعدادات جوالك؟", model.KindBoolean, all, false),
		q(secA11y, "هل تحتاج نظارة للقراءة؟", model.KindBoolean, all, false),
		q(secA11y, "هل تميّز بين الأحمر والأخضر بوضوح؟", model.KindBoolean, all, false),
		q(secA11y, "هل تفضّل الوضع الليلي الداكن؟", model.KindSingleChoice, all, false,
			"دائمًا", "في الليل فقط", "لا أفضّله", "لا فرق عندي"),
		q(secA11y, "هل تعتمد على الأيقونات أم على النصوص؟", model.KindSingleChoice, all, false,
			"الأيقونات", "النصوص", "الاثنان معًا"),
		q(secA11y, "هل تستخدم خاصية قراءة الشاشة بالصوت؟", model.KindBoolean, all, false),
		q(secA11y, "ما الذي يجعل الشاشة تبدو «مزحومة» في نظرك؟", model.KindLongText, all, false),
		q(secA11y, "هل بين زملائك من يقرأ ويكتب بصعوبة؟", model.KindBoolean, field, false),
	}
}

// trustQuestions الثقة والخصوصية — عامل تبنٍّ حاسم في التتبع الميداني.
func trustQuestions() []model.Question {
	return []model.Question{
		q(secTrust, "كيف تشعر تجاه تتبّع موقعك أثناء الوردية؟", model.KindSingleChoice, staff, false,
			"طبيعي ومقبول", "مقبول مع تحفّظ", "يزعجني", "أرفضه"),
		q(secTrust, "هل يهمك أن يتوقف التتبّع بعد نهاية الوردية؟", model.KindBoolean, staff, false),
		q(secTrust, "ما المعلومات التي لا تحب أن تراها الشركة عنك؟", model.KindLongText, staff, false),
		q(secTrust, "هل تخشى أن يُستخدم التطبيق ضدك؟ وضّح.", model.KindLongText, staff, false),
		q(secTrust, "ما الذي يجعلك تثق بأن التطبيق منصف؟", model.KindLongText, staff, false),
		q(secTrust, "هل تحب أن يرى زملاؤك تقييمك أو ترتيبك؟", model.KindBoolean, staff, false),
		q(secTrust, "ما رأيك في نقاط ومكافآت على الأداء داخل التطبيق؟", model.KindLongText, staff, false),
	}
}

// trainQuestions التدريب والتبنّي.
func trainQuestions() []model.Question {
	return []model.Question{
		q(secTrain, "كيف تعلّمت الأنظمة التي تستخدمها اليوم؟", model.KindSingleChoice, all, false,
			"تدريب رسمي", "زميل شرح لي", "جرّبت بنفسي", "لم أتعلّمها فعلًا"),
		q(secTrain, "كم يحتاج موظف جديد ليتقن النظام؟", model.KindSingleChoice, all, false,
			"يوم واحد", "أسبوع", "شهر", "أكثر"),
		q(secTrain, "ما الذي يجعل زميلك يرفض استخدام التطبيق؟", model.KindLongText, all, false),
		q(secTrain, "هل تحتاج شرحًا داخل التطبيق أول مرة؟", model.KindBoolean, all, false),
		q(secTrain, "أي شرح تفضّل؟", model.KindSingleChoice, all, false,
			"فيديو قصير", "صور بخطوات", "نص مكتوب", "لا أحتاج شرحًا"),
		q(secTrain, "من تسأل عندما لا تعرف شيئًا في النظام؟", model.KindShortText, all, false),
	}
}

// teamQuestions الإشراف على الفريق — رحلة J-SUP-01/02.
func teamQuestions() []model.Question {
	return []model.Question{
		q(secTeam, "كم موظفًا تحت إشرافك؟", model.KindSingleChoice, sup, false,
			"أقل من ١٠", "من ١٠ إلى ٣٠", "من ٣٠ إلى ١٠٠", "أكثر من ١٠٠"),
		q(secTeam, "كم موقعًا تزور في اليوم؟", model.KindSingleChoice, sup, false,
			"موقع واحد", "من ٢ إلى ٤", "من ٥ إلى ١٠", "أكثر من ١٠"),
		q(secTeam, "كيف تعرف الآن أن الحارس في موقعه؟", model.KindLongText, sup, false),
		q(secTeam, "كيف تتأكد أن الجولات نُفّذت فعلًا لا شكليًا؟", model.KindLongText, sup, false),
		q(secTeam, "كيف تتعامل مع غياب مفاجئ؟", model.KindLongText, sup, false),
		q(secTeam, "كيف توزّع الموظفين على المواقع والورديات؟", model.KindLongText, sup, false),
		q(secTeam, "كم من وقتك اليومي يذهب في متابعة الحضور؟", model.KindSingleChoice, sup, false,
			"أقل من نصف ساعة", "من ٣٠ إلى ٦٠ دقيقة", "من ١ إلى ٣ ساعات", "أكثر"),
		q(secTeam, "ما القرار الذي تحتاج اتخاذه بسرعة وأنت في الميدان؟", model.KindLongText, sup, false),
		q(secTeam, "ما الذي تكتشفه متأخرًا وتتمنى معرفته فورًا؟", model.KindLongText, sup, false),
		q(secTeam, "ما أكثر شكوى تصلك من الحراس؟", model.KindLongText, sup, false),
		q(secTeam, "هل تنشئ مهامًا أو تعاميم للفريق؟ كيف؟", model.KindLongText, sup, false),
		q(secTeam, "هل تستطيع إنجاز عملك الإشرافي من الجوال وحده؟", model.KindLongText, sup, false),
	}
}

// reviewQuestions المراجعة والاعتماد — رحلة J-SUP-03.
func reviewQuestions() []model.Question {
	return []model.Question{
		q(secReview, "ما الذي تعتمده أو ترفضه يوميًا؟", model.KindLongText, sup, false),
		q(secReview, "كم طلبًا أو بلاغًا تراجع في اليوم؟", model.KindSingleChoice, sup, false,
			"أقل من ٥", "من ٥ إلى ٢٠", "من ٢٠ إلى ٥٠", "أكثر من ٥٠"),
		q(secReview, "ما المعلومات التي تحتاجها لتقرر خلال ثوانٍ؟", model.KindLongText, sup, false),
		q(secReview, "متى تحتاج فتح التفاصيل كاملة قبل القرار؟", model.KindLongText, sup, false),
		q(secReview, "هل تعتمد طلبات كثيرة دفعة واحدة؟", model.KindBoolean, sup, false),
		q(secReview, "ماذا يحدث لو اعتمدت شيئًا بالخطأ؟", model.KindLongText, sup, false),
		q(secReview, "كيف تعرف أن قرارك وصل صاحب الطلب؟", model.KindLongText, sup, false),
		q(secReview, "ما الذي يتراكم عليك ولا تجد وقتًا لمراجعته؟", model.KindLongText, sup, false),
	}
}

// gsQuestions الإشراف العام والخريطة — رحلة J-GS-01.
func gsQuestions() []model.Question {
	return []model.Question{
		q(secGS, "كم موقعًا تحت مسؤوليتك؟", model.KindSingleChoice, mgmt, false,
			"أقل من ٥", "من ٥ إلى ٢٠", "من ٢٠ إلى ٥٠", "أكثر من ٥٠"),
		q(secGS, "كيف تقارن أداء موقع بموقع آخر؟", model.KindLongText, mgmt, false),
		q(secGS, "ما المؤشر الذي يخبرك أن موقعًا فيه مشكلة؟", model.KindLongText, mgmt, false),
		q(secGS, "هل تحتاج رؤية مواقع الفرق على خريطة مباشرة؟", model.KindBoolean, mgmt, false),
		q(secGS, "ماذا تفعل بالمعلومة التي تراها على الخريطة؟", model.KindLongText, mgmt, false),
		q(secGS, "كم دقيقة تأخير في تحديث الموقع تعتبرها مقبولة؟", model.KindSingleChoice, mgmt, false,
			"لحظي", "أقل من دقيقة", "من ١ إلى ٥ دقائق", "لا يهم"),
		q(secGS, "كيف تسجّل ملاحظاتك على موقع بعد زيارته؟", model.KindLongText, mgmt, false),
		q(secGS, "كيف تخطّط تغطية الإجازات والغياب؟", model.KindLongText, mgmt, false),
		q(secGS, "ما الذي يستهلك وقتك أكثر في إدارة المواقع؟", model.KindLongText, mgmt, false),
	}
}

// webDailyQuestions لوحة الويب — الاستخدام اليومي. مدير الشركة مكتبي لا ميداني.
func webDailyQuestions() []model.Question {
	return []model.Question{
		q(secWebDaily, "على أي جهاز تفتح لوحة الإدارة؟", model.KindMultiChoice, manager, false,
			"لابتوب", "كمبيوتر مكتبي", "آيباد", "جوال"),
		q(secWebDaily, "كم ساعة تقضيها أمام اللوحة يوميًا؟", model.KindSingleChoice, manager, false,
			"أقل من ساعة", "من ١ إلى ٣ ساعات", "من ٣ إلى ٦ ساعات", "أكثر من ٦"),
		q(secWebDaily, "كم شاشة أو نافذة تفتح في وقت واحد؟", model.KindSingleChoice, manager, false,
			"واحدة", "اثنتان", "ثلاث فأكثر"),
		q(secWebDaily, "ما أول شاشة تريد أن تفتح عليها اللوحة؟", model.KindLongText, manager, true),
		q(secWebDaily, "ما الذي تراقبه طوال اليوم دون أن تغلقه؟", model.KindLongText, manager, false),
		q(secWebDaily, "كم مؤشرًا تحتاج رؤيته دفعة واحدة؟", model.KindSingleChoice, manager, false,
			"من ٣ إلى ٤", "من ٥ إلى ٨", "أكثر من ٨"),
		q(secWebDaily, "ما الذي يجب أن ينبّهك فورًا وأنت في شاشة أخرى؟", model.KindLongText, manager, false),
		q(secWebDaily, "هل تعمل على اللوحة أثناء اجتماعات أو مكالمات؟", model.KindBoolean, manager, false),
		q(secWebDaily, "ما الذي يرهق عينك بعد ساعات أمام الشاشة؟", model.KindLongText, manager, false),
		q(secWebDaily, "هل تستخدم اختصارات لوحة المفاتيح؟", model.KindBoolean, manager, false),
	}
}

// webDataQuestions لوحة الويب — الجداول والتقارير.
func webDataQuestions() []model.Question {
	return []model.Question{
		q(secWebData, "ما الجدول الذي تفتحه أكثر من غيره؟", model.KindLongText, manager, false),
		q(secWebData, "كم صفًا تتحمّل تصفّحه قبل أن تمل؟", model.KindSingleChoice, manager, false,
			"أقل من ٢٠", "من ٢٠ إلى ٥٠", "من ٥٠ إلى ٢٠٠", "لا يهم إن كان البحث جيدًا"),
		q(secWebData, "بأي شيء تفلتر عادة؟", model.KindMultiChoice, manager, false,
			"التاريخ", "الموقع", "الموظف", "الحالة", "العميل"),
		q(secWebData, "هل تحتاج حفظ فلاتر تستخدمها باستمرار؟", model.KindBoolean, manager, false),
		q(secWebData, "هل تصدّر البيانات إلى إكسل؟ ولماذا؟", model.KindLongText, manager, false),
		q(secWebData, "ما التقرير الذي تُعدّه يدويًا وتتمنى أن يخرج تلقائيًا؟", model.KindLongText, manager, false),
		q(secWebData, "هل تطبع تقارير على ورق؟", model.KindBoolean, manager, false),
		q(secWebData, "ما الأعمدة التي تنظر إليها أولًا في أي جدول؟", model.KindLongText, manager, false),
		q(secWebData, "هل تحتاج تعديل البيانات من الجدول مباشرة؟", model.KindBoolean, manager, false),
		q(secWebData, "ما الذي يزعجك في أنظمة الويب التي استخدمتها؟", model.KindLongText, manager, false),
	}
}

// financeQuestions المالية والفوترة والعقود.
func financeQuestions() []model.Question {
	return []model.Question{
		q(secFinance, "كيف تُحتسب فاتورة العميل اليوم؟", model.KindLongText, manager, false),
		q(secFinance, "ما الذي يسبب خلافًا على الفاتورة؟", model.KindLongText, manager, false),
		q(secFinance, "كيف تربط ساعات الحراسة الفعلية بالفوترة؟", model.KindLongText, manager, false),
		q(secFinance, "كم يستغرق إصدار فواتير الشهر؟", model.KindSingleChoice, manager, false,
			"ساعات", "يوم أو يومان", "أسبوع", "أكثر"),
		q(secFinance, "ما الذي تتابعه من مستحقات وتحصيل؟", model.KindLongText, manager, false),
		q(secFinance, "كيف تتابع تواريخ انتهاء العقود؟", model.KindLongText, manager, false),
		q(secFinance, "ما التنبيه المالي الذي تحتاجه قبل فوات الأوان؟", model.KindLongText, manager, false),
	}
}

// hrQuestions الموارد البشرية والرواتب.
func hrQuestions() []model.Question {
	return []model.Question{
		q(secHR, "كيف تُحتسب الرواتب اليوم؟", model.KindLongText, mgmt, false),
		q(secHR, "ما الذي يتسبب في أخطاء الرواتب؟", model.KindLongText, mgmt, false),
		q(secHR, "كيف تتابع انتهاء الإقامات والتأمين والرخص؟", model.KindLongText, mgmt, false),
		q(secHR, "كم موظفًا توظّف شهريًا تقريبًا؟", model.KindSingleChoice, mgmt, false,
			"أقل من ٥", "من ٥ إلى ٢٠", "من ٢٠ إلى ٥٠", "أكثر من ٥٠"),
		q(secHR, "ما أطول خطوة في توظيف حارس جديد؟", model.KindLongText, mgmt, false),
		q(secHR, "ما سبب ترك الموظفين للعمل عندكم؟", model.KindLongText, mgmt, false),
		q(secHR, "ما المستندات التي تُطلب من كل موظف؟", model.KindLongText, mgmt, false),
	}
}

// clientQuestions العملاء والقرارات.
func clientQuestions() []model.Question {
	return []model.Question{
		q(secClient, "ما الرقم الذي يسألك عنه العميل دائمًا؟", model.KindLongText, mgmt, false),
		q(secClient, "كيف تثبت للعميل أن الخدمة نُفّذت؟", model.KindLongText, mgmt, false),
		q(secClient, "ما أكثر سبب لشكاوى العملاء؟", model.KindLongText, mgmt, false),
		q(secClient, "هل يحتاج العميل دخولًا على النظام ليرى موقعه؟", model.KindBoolean, mgmt, false),
		q(secClient, "ما القرار الأسبوعي الذي تعتمد فيه على أرقام؟", model.KindLongText, mgmt, false),
		q(secClient, "ما البيانات التي تجمعها يدويًا حتى اليوم؟", model.KindLongText, mgmt, false),
		q(secClient, "ما التقرير الذي ترفعه للإدارة العليا أو الملّاك؟", model.KindLongText, manager, false),
		q(secClient, "كم من وقتك الأسبوعي يذهب في إعداد التقارير؟", model.KindSingleChoice, mgmt, false,
			"أقل من ساعة", "من ١ إلى ٣ ساعات", "من ٣ إلى ٨ ساعات", "أكثر"),
	}
}

// hiddenQuestions الاحتياجات الخفية.
func hiddenQuestions() []model.Question {
	return []model.Question{
		q(secHidden, "ما أكثر شيء تنساه أثناء العمل؟", model.KindLongText, all, false),
		q(secHidden, "ما أكثر شيء تتصل بالمشرف لأجله؟", model.KindLongText, field, false),
		q(secHidden, "ما أكثر سؤال يسألك إياه المشرف؟", model.KindLongText, field, false),
		q(secHidden, "ما أكثر مهمة تكررها كل يوم؟", model.KindLongText, all, false),
		q(secHidden, "لو منعناك من استخدام الورق نهائيًا، ماذا تحتاج داخل التطبيق؟", model.KindLongText, all, false),
		q(secHidden, "ما الأشياء التي تحفظها في رأسك وتتمنى أن يتذكرها التطبيق بدلًا عنك؟", model.KindLongText, all, false),
		q(secHidden, "ما الشيء الذي تفعله خارج النظام لأن النظام لا يدعمه؟", model.KindLongText, all, false),
		q(secHidden, "ما الورقة التي ما زلت تستخدمها رغم وجود الأنظمة؟", model.KindLongText, all, false),
		q(secHidden, "ما المعلومة التي تصوّرها بجوالك لتتذكرها لاحقًا؟", model.KindLongText, all, false),
	}
}

// priorityQuestions الأولويات — أسئلة ترتيب تحسم المفاضلات التصميمية.
func priorityQuestions() []model.Question {
	return []model.Question{
		q(secPriority, "رتّب هذه الميزات من الأكثر أهمية إلى الأقل.", model.KindRanking, all, true,
			"الحضور", "الدوريات", "الخرائط", "البلاغات", "الطلبات", "الإشعارات", "الدردشة", "الملفات"),
		q(secPriority, "رتّب ما يعطّل يومك من الأكثر إلى الأقل.", model.KindRanking, all, false,
			"ضعف الشبكة", "كثرة الكتابة", "الاتصالات والمقاطعات", "التنقّل بين المواقع",
			"بطء الأنظمة", "تعليمات غير واضحة"),
		q(secPriority, "رتّب ما تريد رؤيته في الشاشة الأولى.", model.KindRanking, field, false,
			"حالتي الآن", "جولتي القادمة", "البلاغات المفتوحة", "رسائل المشرف",
			"ورديّاتي القادمة", "طلباتي"),
		q(secPriority, "رتّب أولوياتك في لوحة الويب.", model.KindRanking, mgmt, false,
			"الحضور المباشر", "تقارير الجولات", "البلاغات والحوادث", "أداء المواقع",
			"الفوترة والعقود", "شؤون الموظفين"),
	}
}

// finalQuestions أسئلة ختامية.
func finalQuestions() []model.Question {
	return []model.Question{
		q(secFinal, "ما أكثر مشكلة تواجهها أثناء العمل؟", model.KindLongText, all, true),
		q(secFinal, "ما الذي يجب أن يعمل حتى لو انقطع الإنترنت؟", model.KindLongText, all, false),
		q(secFinal, "لو حذفنا كل شيء من التطبيق، ما الميزة الوحيدة التي لا يمكن الاستغناء عنها؟", model.KindLongText, all, true),
		q(secFinal, "لو كان لك قرار واحد لتغيير التطبيق، ماذا ستغيّر؟", model.KindLongText, all, false),
		q(secFinal, "ما الشيء الذي تتمنى أن يفعله التطبيق ولم تجده في أي نظام استخدمته؟", model.KindLongText, all, false),
		q(secFinal, "ما الذي يجعلك تستخدم التطبيق دون أن يجبرك أحد؟", model.KindLongText, all, false),
		q(secFinal, "هل تريد إضافة شيء لم نسألك عنه؟", model.KindLongText, all, false),
		q(secFinal, "كم تقيّم وضوح هذه الأسئلة؟ (١ = غير واضحة، ٥ = واضحة جدًا)", model.KindScale, all, false),
	}
}
