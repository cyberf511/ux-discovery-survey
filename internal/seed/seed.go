// Package seed يحمّل أسئلة المقابلة الأولية عند أول تشغيل.
package seed

import (
	"fmt"

	"surveyapp/internal/model"
	"surveyapp/internal/store"
)

// الفئات المستخدمة في التوسيم أدناه.
var (
	all   = []model.Category{} // فارغ = يظهر لكل الفئات
	field = []model.Category{model.CatGuard, model.CatSupervisor}
	staff = []model.Category{model.CatGuard, model.CatSupervisor, model.CatAreaManager}
)

// Apply يحمّل الأسئلة الأولية إن لم تكن حُمِّلت من قبل.
func Apply(s *store.Store) (int, error) {
	done, err := s.Seeded()
	if err != nil {
		return 0, err
	}
	if done {
		return 0, nil
	}
	existing, err := s.Questions(true)
	if err != nil {
		return 0, err
	}
	if len(existing) > 0 {
		// قاعدة بيانات فيها أسئلة أصلًا: لا نكرّرها، فقط نعلّمها كمُحمَّلة.
		return 0, s.MarkSeeded()
	}
	n := 0
	for _, q := range Questions() {
		if _, err := s.CreateQuestion(q); err != nil {
			return n, fmt.Errorf("إضافة سؤال %q: %w", q.Text, err)
		}
		n++
	}
	return n, s.MarkSeeded()
}

func q(section, text string, kind model.Kind, cats []model.Category, required bool, opts ...string) model.Question {
	if opts == nil {
		opts = []string{}
	}
	return model.Question{
		Text:       text,
		Kind:       kind,
		Section:    section,
		Categories: cats,
		Options:    opts,
		Required:   required,
	}
}

// Questions قائمة أسئلة مقابلة UX Discovery الأولية بالترتيب المعروض.
func Questions() []model.Question {
	const (
		secDay      = "فهم يوم العمل"
		secHome     = "الشاشة الرئيسية"
		secAttend   = "الحضور والانصراف"
		secPatrol   = "الدوريات"
		secCheck    = "نقاط التفتيش"
		secReport   = "البلاغات"
		secRequest  = "الطلبات"
		secNotify   = "الإشعارات"
		secPerf     = "الأداء"
		secPriority = "الأولويات"
		secHidden   = "الاحتياجات الخفية"
		secFinal    = "أسئلة ختامية"
	)

	return []model.Question{
		// فهم يوم العمل
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

		// الشاشة الرئيسية
		q(secHome, "عندما تفتح التطبيق، ما أول معلومة تريد رؤيتها؟", model.KindLongText, all, true),
		q(secHome, "ما أهم ثلاث معلومات بالنسبة لك؟", model.KindLongText, all, false),
		q(secHome, "ما الأشياء التي لا تحتاج رؤيتها كل مرة؟", model.KindLongText, all, false),
		q(secHome, "إذا فتحت التطبيق لمدة خمس ثوانٍ فقط، ماذا يجب أن تعرف؟", model.KindLongText, all, false),

		// الحضور والانصراف
		q(secAttend, "كيف تسجل حضورك الآن؟", model.KindLongText, field, false),
		q(secAttend, "ما المشاكل التي تواجهها في تسجيل الحضور؟", model.KindLongText, field, false),
		q(secAttend, "هل يحدث أن يرفض تسجيل الحضور؟", model.KindBoolean, field, false),
		q(secAttend, "كيف تعرف أن حضورك تم بنجاح؟", model.KindLongText, field, false),
		q(secAttend, "إذا فشل التسجيل، ماذا تتوقع أن يحدث؟", model.KindLongText, field, false),

		// الدوريات
		q(secPatrol, "كيف تعرف أين تذهب في الجولة؟", model.KindLongText, field, false),
		q(secPatrol, "كيف تعرف أنك وصلت؟", model.KindLongText, field, false),
		q(secPatrol, "كيف تعرف أنك انتهيت من الجولة؟", model.KindLongText, field, false),
		q(secPatrol, "ماذا تحتاج أثناء المشي في الجولة؟", model.KindLongText, field, false),
		q(secPatrol, "هل تحتاج الخريطة طوال الوقت؟", model.KindBoolean, field, false),
		q(secPatrol, "ماذا لو انقطع الإنترنت أثناء الجولة؟", model.KindLongText, field, false),
		q(secPatrol, "ماذا لو ضعت عن المسار؟", model.KindLongText, field, false),

		// نقاط التفتيش
		q(secCheck, "كيف تثبت أنك وصلت للنقطة؟", model.KindLongText, field, false),
		q(secCheck, "أي وسيلة إثبات وصول تفضّل؟", model.KindSingleChoice, field, false,
			"QR", "NFC", "GPS", "لا يهم"),
		q(secCheck, "هل سبق أن واجهت مشكلة في إثبات الوصول؟ صفها.", model.KindLongText, field, false),
		q(secCheck, "ما أكثر شيء يبطئك عند نقاط التفتيش؟", model.KindLongText, field, false),

		// البلاغات
		q(secReport, "متى تنشئ بلاغًا؟", model.KindLongText, all, false),
		q(secReport, "ما أنواع البلاغات الأكثر شيوعًا عندك؟", model.KindLongText, all, false),
		q(secReport, "كيف تفضّل تسجيل البلاغ؟", model.KindSingleChoice, all, false,
			"الكتابة", "الصوت", "الصور", "مزيج بينها"),
		q(secReport, "ما المعلومات التي يجب أن تُطلب منك في البلاغ؟", model.KindLongText, all, false),
		q(secReport, "ما المعلومات التي لا ينبغي أن تُطلب منك؟", model.KindLongText, all, false),

		// الطلبات
		q(secRequest, "ما الطلبات التي تقدمها غالبًا؟ (إجازة، استئذان، سلفة...)", model.KindLongText, staff, false),
		q(secRequest, "ما المعلومات المطلوبة لكل طلب؟", model.KindLongText, staff, false),
		q(secRequest, "كيف تعرف أن الطلب وصل؟", model.KindLongText, staff, false),
		q(secRequest, "ماذا تريد أن ترى بعد إرسال الطلب؟", model.KindLongText, staff, false),

		// الإشعارات
		q(secNotify, "ما الإشعارات المهمة بالنسبة لك؟", model.KindLongText, all, false),
		q(secNotify, "متى تعتبر الإشعار مزعجًا؟", model.KindLongText, all, false),
		q(secNotify, "كيف تميّز بين الإشعار العادي والعاجل؟", model.KindLongText, all, false),
		q(secNotify, "ما الذي يجب أن يظهر في الإشعار دون الحاجة لفتح التطبيق؟", model.KindLongText, all, false),

		// الأداء
		q(secPerf, "كم ثانية تعتبر انتظارًا مقبولًا؟", model.KindSingleChoice, all, false,
			"أقل من ثانية", "من ١ إلى ٣ ثوانٍ", "من ٣ إلى ٥ ثوانٍ", "أكثر من ٥ ثوانٍ"),
		q(secPerf, "إذا كان الإنترنت ضعيفًا، ماذا تتوقع من التطبيق؟", model.KindLongText, all, false),
		q(secPerf, "هل تفضّل أن يعمل التطبيق بدون إنترنت؟", model.KindBoolean, all, false),
		q(secPerf, "ما أسوأ شيء يمكن أن يحدث أثناء العمل؟", model.KindLongText, all, false),

		// الأولويات
		q(secPriority, "رتّب هذه الميزات من الأكثر أهمية إلى الأقل.", model.KindRanking, all, true,
			"الحضور", "الدوريات", "الخرائط", "البلاغات", "الطلبات", "الإشعارات", "الدردشة", "الملفات"),

		// الاحتياجات الخفية
		q(secHidden, "ما أكثر شيء تنساه أثناء العمل؟", model.KindLongText, all, false),
		q(secHidden, "ما أكثر شيء تتصل بالمشرف لأجله؟", model.KindLongText, field, false),
		q(secHidden, "ما أكثر سؤال يسألك إياه المشرف؟", model.KindLongText, field, false),
		q(secHidden, "ما أكثر مهمة تكررها كل يوم؟", model.KindLongText, all, false),
		q(secHidden, "لو منعناك من استخدام الورق نهائيًا، ماذا تحتاج داخل التطبيق؟", model.KindLongText, all, false),
		q(secHidden, "ما الأشياء التي تحفظها في رأسك وتتمنى أن يتذكرها التطبيق بدلًا عنك؟", model.KindLongText, all, false),

		// أسئلة ختامية
		q(secFinal, "ما أكثر مشكلة تواجهها أثناء العمل؟", model.KindLongText, all, true),
		q(secFinal, "ما الذي يجب أن يعمل حتى لو انقطع الإنترنت؟", model.KindLongText, all, false),
		q(secFinal, "لو حذفنا كل شيء من التطبيق، ما الميزة الوحيدة التي لا يمكن الاستغناء عنها؟", model.KindLongText, all, true),
		q(secFinal, "لو كان لك قرار واحد لتغيير التطبيق، ماذا ستغيّر؟", model.KindLongText, all, false),
		q(secFinal, "ما الشيء الذي تتمنى أن يفعله التطبيق ولم تجده في أي نظام استخدمته؟", model.KindLongText, all, false),
		q(secFinal, "كم تقيّم وضوح هذه الأسئلة؟ (١ = غير واضحة، ٥ = واضحة جدًا)", model.KindScale, all, false),
	}
}
