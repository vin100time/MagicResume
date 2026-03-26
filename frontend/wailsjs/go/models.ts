export namespace main {
	
	export class Recording {
	    id: string;
	    name: string;
	    date: string;
	    duration: number;
	    status: string;
	    audioPath: string;
	    transcriptPath?: string;
	    summaryPath?: string;
	
	    static createFrom(source: any = {}) {
	        return new Recording(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.date = source["date"];
	        this.duration = source["duration"];
	        this.status = source["status"];
	        this.audioPath = source["audioPath"];
	        this.transcriptPath = source["transcriptPath"];
	        this.summaryPath = source["summaryPath"];
	    }
	}
	export class Settings {
	    promptTemplate: string;
	    model: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.promptTemplate = source["promptTemplate"];
	        this.model = source["model"];
	    }
	}

}

